package funnel

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

func (f *Funnel) handleServerWebSocket(serverId string, c *gin.Context) {
	qq.Println("@Funnel/handleServerWebSocket/1{SERVER_ID}", serverId)
	// Upgrade to websocket
	conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		qq.Println("@Funnel/handleServerWebSocket/3{ERROR}", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade websocket"})
		return
	}

	qq.Println("@Funnel/handleServerWebSocket/2{CONN}", conn)

	// Register the server connection
	f.registerServer(serverId, conn)
}

func (f *Funnel) registerServer(serverId string, conn net.Conn) {
	qq.Println("@Funnel/registerServer/1{SERVER_ID}", serverId)
	f.scLock.Lock()
	f.serverConnections[serverId] = &ServerConnection{Conn: conn}
	f.scLock.Unlock()

	// Start goroutine to handle incoming responses from this server
	go f.handleServerConnection(serverId, conn)
}

// handleServerConnection handles incoming packets from a server connection
func (f *Funnel) handleServerConnection(serverId string, conn net.Conn) {
	qq.Println("@Funnel/handleServerConnection/1{SERVER_ID}", serverId)

	defer conn.Close()

	// Read request ID (16 bytes) first, then packet
	reqIdBuf := make([]byte, 16)

	for {
		qq.Println("@Funnel/handleServerConnection/2{REQ_ID_BUF_READ}")
		// Read request ID
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			if err != io.EOF {
				// Connection closed or error
			}
			break
		}

		reqId := kosher.Str(reqIdBuf)

		qq.Println("@Funnel/handleServerConnection/3{REQ_ID}", reqId)

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

		// Skip WebSocket data packets (they're handled in routeWS)
		if headerPacket.PType == PtypeWebSocketData {
			continue
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

		// Read all body packets before sending response
		// This ensures we don't have a race condition with responseReader
		var bodyData []byte
		if headerPacket.Total > 0 {
			// Read body packets
			remaining := int64(headerPacket.Total)
			for remaining > 0 {
				bodyPacket, err := ReadPacket(conn)
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

				if bodyPacket.PType != PtypeSendBody && bodyPacket.PType != PtypeEndBody {
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

				bodyData = append(bodyData, bodyPacket.Data...)
				remaining -= int64(len(bodyPacket.Data))

				if bodyPacket.PType == PtypeEndBody {
					break
				}
			}
		} else if headerPacket.Total == 0 {
			// Read EndBody packet for zero-length response
			endBodyPacket, err := ReadPacket(conn)
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
			if endBodyPacket.PType != PtypeEndBody {
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
		}

		// Create response body from buffered data
		resp.Body = io.NopCloser(bytes.NewReader(bodyData))

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
