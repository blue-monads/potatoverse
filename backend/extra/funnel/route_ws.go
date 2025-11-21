package funnel

import (
	"errors"
	"io"
	"net/http/httputil"

	"github.com/blue-monads/turnix/backend/utils/qq"
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
		qq.Println("@routeWS/1{SERVER_NOT_CONNECTED}")
		c.Error(errors.New("server not connected"))
		return
	}

	// Generate request ID
	reqId := GetRequestId()

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		qq.Println("@routeWS/2{ERROR}", err)
		c.Error(err)
		return
	}

	pendingReqChan := make(chan *Packet)
	f.pendingReqLock.Lock()
	f.pendingReq[reqId] = pendingReqChan
	f.pendingReqLock.Unlock()

	defer func() {
		qq.Println("@cleanup/1{REQ_ID}", reqId)
		f.pendingReqLock.Lock()
		delete(f.pendingReq, reqId)
		f.pendingReqLock.Unlock()
	}()

	// Write request header packet
	serverConn.writeChan <- &ServerWrite{
		packet: &Packet{
			PType:  PTypeSendHeader,
			Offset: 0,
			Total:  0, // WebSocket doesn't have a body in the initial request
			Data:   out,
		},
		reqId: reqId,
	}

	// Upgrade client connection to websocket
	clientConn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		qq.Println("@upgrade_err", err)
		c.Error(err)
		return
	}
	defer clientConn.Close()

	go func() {
		for {

			packet := <-pendingReqChan
			if packet == nil {
				break
			}

			err = wsutil.WriteServerBinary(clientConn, packet.Data)
			if err != nil {
				break
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

		// Write WebSocket data as packet
		serverConn.writeChan <- &ServerWrite{
			packet: &Packet{
				PType:  PtypeWebSocketData,
				Offset: 0,
				Total:  int32(len(msg)),
				Data:   msg,
			},
			reqId: reqId,
		}

		// If it's a close message, break
		if op == ws.OpClose {
			break
		}
	}
}
