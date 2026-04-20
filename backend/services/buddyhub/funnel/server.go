package funnel

import (
	"io"
	"net"
	"net/http"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub/packetwire"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

func (f *Funnel) handleServerWebSocket(nodeId string, c *gin.Context) {
	qq.Println("@Funnel/handleServerWebSocket/1{SERVER_ID}", nodeId)
	// Upgrade to websocket
	conn, _, _, err := ws.UpgradeHTTP(c.Request, c.Writer)
	if err != nil {
		qq.Println("@Funnel/handleServerWebSocket/3{ERROR}", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upgrade websocket"})
		return
	}

	qq.Println("@Funnel/handleServerWebSocket/2{CONN}")

	// Register the server connection
	f.registerServer(nodeId, conn)
}

func (f *Funnel) registerServer(nodeId string, conn net.Conn) {
	qq.Println("@Funnel/registerServer/1{SERVER_ID}", nodeId)
	f.scLock.Lock()

	pool, exists := f.serverPools[nodeId]
	if !exists {
		pool = &ServerPool{
			handles: []*ServerHandle{},
		}
		f.serverPools[nodeId] = pool
	}
	f.scLock.Unlock()

	swchan := make(chan *ServerWrite)
	handle := &ServerHandle{
		conn:      conn,
		writeChan: swchan,
		nodeId:    nodeId,
	}

	pool.lock.Lock()
	pool.handles = append(pool.handles, handle)
	pool.lock.Unlock()

	// Start goroutine to handle incoming responses from this server
	go f.handleServerConnection(handle, true, func() {
		pool.lock.Lock()
		defer pool.lock.Unlock()
		for i, h := range pool.handles {
			if h == handle {
				pool.handles = append(pool.handles[:i], pool.handles[i+1:]...)
				break
			}
		}

		if len(pool.handles) == 0 {
			f.scLock.Lock()
			if f.serverPools[nodeId] == pool {
				delete(f.serverPools, nodeId)
			}
			f.scLock.Unlock()
		}
	})
}

// handleServerConnection handles incoming packets from a server connection
func (f *Funnel) handleServerConnection(handle *ServerHandle, isWS bool, onExit func()) {
	nodeId := handle.nodeId
	conn := handle.conn
	swchan := handle.writeChan
	qq.Println("@handleServerConnection/1", nodeId)
	defer func() {
		conn.Close()
		close(swchan)

		qq.Println("@handleServerConnection/2", nodeId)
		if onExit != nil {
			onExit()
		}
	}()

	go func() {
		for {
			sw := <-swchan
			if sw == nil {
				break
			}

			err := packetwire.WritePacketFull(conn, sw.packet, sw.reqId)
			if err != nil {
				qq.Println("@handleServerConnection/5{ERROR}", err)
				break
			}

			qq.Println("@write", isWS)

		}
	}()

	for {

		reqIdBuf := make([]byte, 16)
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			qq.Println("@handleServerConnection/3", nodeId, err)
			break
		}

		reqId := string(reqIdBuf)

		qq.Println("@handleServerConnection/4{REQ_ID}", reqId, isWS)

		packet, err := packetwire.ReadPacket(conn)
		if err != nil {
			qq.Println("@handleServerConnection/3", nodeId, err)
			break
		}

		f.pendingReqLock.Lock()
		pendingReqChan := f.pendingReq[reqId]
		f.pendingReqLock.Unlock()

		if pendingReqChan == nil {
			qq.Println("@handleServerConnection/5{PENDING_REQ_NOT_FOUND}", reqId, "PACKET_TYPE", packet.PType)
			continue
		}

		qq.Println("@handleServerConnection/6{ROUTING_PACKET}", reqId, "PACKET_TYPE", packet.PType, "DATA_LEN", len(packet.Data))
		pendingReqChan <- packet
	}

}
