package funnel

import (
	"errors"
	"io"
	"net/http/httputil"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func (f *Funnel) routeWS(serverId string, c *gin.Context) {

	qq.Println("@routeWS/1", serverId)

	f.scLock.RLock()
	serverConn, exists := f.serverConnections[serverId]
	f.scLock.RUnlock()

	if !exists {
		qq.Println("@routeWS/1{SERVER_NOT_CONNECTED}")
		c.Error(errors.New("server not connected"))
		return
	}

	qq.Println("@routeWS/2")

	// Generate request ID
	reqId := GetRequestId()

	qq.Println("@routeWS/3")

	// Dump request
	req := c.Request
	out, err := httputil.DumpRequest(req, false)
	if err != nil {
		qq.Println("@routeWS/2{ERROR}", err)
		c.Error(err)
		return
	}

	qq.Println("@routeWS/4")

	pendingReqChan := make(chan *Packet)
	f.pendingReqLock.Lock()
	f.pendingReq[reqId] = pendingReqChan
	f.pendingReqLock.Unlock()

	qq.Println("@routeWS/5")

	defer func() {
		qq.Println("@cleanup/1{REQ_ID}", reqId)
		f.pendingReqLock.Lock()
		delete(f.pendingReq, reqId)
		f.pendingReqLock.Unlock()
	}()

	qq.Println("@routeWS/6")

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

	qq.Println("@routeWS/7")

	// Upgrade client connection to websocket
	clientConn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		qq.Println("@upgrade_err", err)
		c.Error(err)
		return
	}

	qq.Println("@routeWS/8")

	defer clientConn.Close()

	qq.Println("@routeWS/9")

	go func() {
		for {

			qq.Println("@routeWS/10/loop")

			packet := <-pendingReqChan
			if packet == nil {
				qq.Println("@routeWS/11/loop/break")
				break
			}

			qq.Println("@routeWS/12/loop/write")

			err = wsutil.WriteServerBinary(clientConn, packet.Data)
			if err != nil {
				qq.Println("@routeWS/13/loop/write/break", err)
				break
			}

		}
	}()

	qq.Println("@routeWS/14")

	for {

		qq.Println("@routeWS/15/loop")

		msg, op, err := wsutil.ReadClientData(clientConn)
		if err != nil {
			if err != io.EOF {
				qq.Println("@routeWS/16/loop/break", err)
				// Connection closed
			}
			break
		}

		qq.Println("@routeWS/17/loop/write")

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

		qq.Println("@routeWS/18/loop/write/end")

		// If it's a close message, break
		if op == ws.OpClose {
			qq.Println("@routeWS/19/loop/break/close")
			break
		}
	}
}
