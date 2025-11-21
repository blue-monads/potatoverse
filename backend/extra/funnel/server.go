package funnel

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

func (f *Funnel) handleServerWebSocket(serverId string, c *gin.Context) {
	// Upgrade to websocket
	conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade websocket"})
		return
	}

	// Register the server connection
	f.registerServer(serverId, conn)
}

func (f *Funnel) registerServer(serverId string, conn net.Conn) {
	f.scLock.Lock()
	f.serverConnections[serverId] = &ServerConnection{Conn: conn}
	f.scLock.Unlock()

	// Start goroutine to handle incoming responses from this server
	go f.handleServerConnection(serverId, conn)
}

// handleServerConnection handles incoming packets from a server connection
func (f *Funnel) handleServerConnection(serverId string, conn net.Conn) {
	defer conn.Close()

	// Read request ID (16 bytes) first, then packet
	reqIdBuf := make([]byte, 16)

	for {
		// Read request ID
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			if err != io.EOF {
				// Connection closed or error
			}
			break
		}

		reqId := kosher.Str(reqIdBuf)

		// Read header packet
		headerPacket, err := ReadPacket(conn)
		if err != nil {
			// Send error to pending request if exists
			f.prLock.Lock()
			if pending, exists := f.pendingRequests[reqId]; exists {
				select {
				case pending.ErrorChan <- err:
				default:
				}
			}
			f.prLock.Unlock()
			break
		}

		if headerPacket.PType != PTypeSendHeader {
			// Invalid packet type
			f.prLock.Lock()
			if pending, exists := f.pendingRequests[reqId]; exists {
				select {
				case pending.ErrorChan <- io.ErrUnexpectedEOF:
				default:
				}
			}
			f.prLock.Unlock()
			continue
		}

		// Parse response header
		reader := bytes.NewBuffer(headerPacket.Data)
		resp, err := http.ReadResponse(bufio.NewReader(reader), nil)
		if err != nil {
			f.prLock.Lock()
			if pending, exists := f.pendingRequests[reqId]; exists {
				select {
				case pending.ErrorChan <- err:
				default:
				}
			}
			f.prLock.Unlock()
			continue
		}

		// Create response reader for body that reads packets directly from connection
		if resp.ContentLength > 0 {
			resp.Body = &responseReader{
				conn:     conn,
				total:    int64(headerPacket.Total),
				received: 0,
			}
		} else {
			// Empty body
			resp.Body = io.NopCloser(bytes.NewReader(nil))
		}

		// Send response to pending request
		f.prLock.Lock()
		if pending, exists := f.pendingRequests[reqId]; exists {
			select {
			case pending.ResponseChan <- resp:
			default:
			}
		}
		f.prLock.Unlock()
	}

	// Clean up connection
	f.scLock.Lock()
	delete(f.serverConnections, serverId)
	f.scLock.Unlock()
}
