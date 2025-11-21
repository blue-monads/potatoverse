package funnel

import (
	"io"
	"net"
	"net/http"

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

	qq.Println("@Funnel/handleServerWebSocket/2{CONN}")

	// Register the server connection
	f.registerServer(serverId, conn)
}

func (f *Funnel) registerServer(serverId string, conn net.Conn) {
	qq.Println("@Funnel/registerServer/1{SERVER_ID}", serverId)
	f.scLock.Lock()
	f.serverConnections[serverId] = conn
	f.scLock.Unlock()

	// Start goroutine to handle incoming responses from this server
	go f.handleServerConnection(serverId, conn)
}

// handleServerConnection handles incoming packets from a server connection
func (f *Funnel) handleServerConnection(serverId string, conn net.Conn) {
	qq.Println("@handleServerConnection/1", serverId)
	defer func() {
		conn.Close()

		qq.Println("@handleServerConnection/2", serverId)
		f.scLock.Lock()
		delete(f.serverConnections, serverId)
		f.scLock.Unlock()
	}()

	for {

		reqIdBuf := make([]byte, 16)
		_, err := io.ReadFull(conn, reqIdBuf)
		if err != nil {
			qq.Println("@handleServerConnection/3", serverId, err)
			break
		}

		reqId := string(reqIdBuf)

		qq.Println("@handleServerConnection/4{REQ_ID}", reqId)

		packet, err := ReadPacket(conn)
		if err != nil {
			qq.Println("@handleServerConnection/3", serverId, err)
			break
		}

		f.pendingReqLock.Lock()
		pendingReqChan := f.pendingReq[reqId]
		f.pendingReqLock.Unlock()

		if pendingReqChan == nil {
			break
		}

		pendingReqChan <- packet
	}

}
