package funnel

import (
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func (f *Funnel) routeWS(serverId string, c *gin.Context) {
	// Get server connection
	f.scLock.RLock()
	serverConn, exists := f.serverConnections[serverId]
	f.scLock.RUnlock()

	if !exists {
		c.JSON(http.StatusBadGateway, gin.H{"error": "server not connected"})
		return
	}

	// Generate request ID
	reqId := GetRequestId()
	reqIdBytes := []byte(reqId)

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Write request ID
	_, err = serverConn.Conn.Write(reqIdBytes)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request id"})
		return
	}

	// Write request header packet
	err = WritePacket(serverConn.Conn, &Packet{
		PType:  PTypeSendHeader,
		Offset: 0,
		Total:  0, // WebSocket doesn't have a body in the initial request
		Data:   out,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to write request header"})
		return
	}

	// Upgrade client connection to websocket
	clientConn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade websocket"})
		return
	}
	defer clientConn.Close()

	// After sending the header packet, websocket communication uses packets with request ID
	// Bidirectionally forward messages between client and server
	// Forward from server to client
	go func() {
		for {
			// Read request ID first
			reqIdBuf := make([]byte, 16)
			_, err := io.ReadFull(serverConn.Conn, reqIdBuf)
			if err != nil {
				if err != io.EOF {
					// Connection error
				}
				clientConn.Close()
				return
			}

			// Verify this is for our request
			if string(reqIdBuf) != reqId {
				// This message is for a different request, skip it
				// Read the packet to consume it
				packet, err := ReadPacket(serverConn.Conn)
				if err != nil {
					clientConn.Close()
					return
				}
				// Skip if not WebSocket data
				if packet.PType != PtypeWebSocketData {
					continue
				}
				// This shouldn't happen, but if it does, we've consumed the packet
				continue
			}

			// Read WebSocket data packet
			packet, err := ReadPacket(serverConn.Conn)
			if err != nil {
				clientConn.Close()
				return
			}

			if packet.PType != PtypeWebSocketData {
				// Invalid packet type, close connection
				clientConn.Close()
				return
			}

			// Write to client websocket as binary
			err = wsutil.WriteServerBinary(clientConn, packet.Data)
			if err != nil {
				clientConn.Close()
				return
			}
		}
	}()

	// Forward from client to server
	for {
		msg, op, err := wsutil.ReadClientData(clientConn)
		if err != nil {
			if err != io.EOF {
				// Connection closed
			}
			break
		}

		// Write request ID
		_, err = serverConn.Conn.Write(reqIdBytes)
		if err != nil {
			break
		}

		// Write WebSocket data as packet
		err = WritePacket(serverConn.Conn, &Packet{
			PType:  PtypeWebSocketData,
			Offset: 0,
			Total:  int32(len(msg)),
			Data:   msg,
		})
		if err != nil {
			break
		}

		// If it's a close message, break
		if op == ws.OpClose {
			break
		}
	}
}
