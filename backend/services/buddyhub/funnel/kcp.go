package funnel

import (
	"io"
	"net"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/xtaci/kcp-go/v5"
)

func (f *Funnel) StartKcpServer() error {
	l, err := kcp.ListenWithOptions(":0", nil, 10, 3)
	if err != nil {
		return err
	}

	f.kcpListener = l
	f.kcpPort = l.Addr().(*net.UDPAddr).Port

	qq.Println("@Funnel/StartKcpServer/1{PORT}", f.kcpPort)

	go f.handleKcpConnections()

	return nil
}

func (f *Funnel) handleKcpConnections() {
	for {
		conn, err := f.kcpListener.Accept()
		if err != nil {
			qq.Println("@Funnel/handleKcpConnections/1{ERROR}", err)
			break
		}

		go f.handleKcpSession(conn)
	}
}

func (f *Funnel) handleKcpSession(conn net.Conn) {
	defer conn.Close()

	// Initial handshake: Read token (64 bytes)
	tokenBuf := make([]byte, 64)
	_, err := io.ReadFull(conn, tokenBuf)
	if err != nil {
		qq.Println("@Funnel/handleKcpSession/1{ERROR}", err)
		return
	}

	// Trim null bytes
	token := ""
	for i := 0; i < len(tokenBuf); i++ {
		if tokenBuf[i] == 0 {
			token = string(tokenBuf[:i])
			break
		}
	}
	if token == "" {
		token = string(tokenBuf)
	}

	nodeId := token // Currently token is nodeId

	qq.Println("@Funnel/handleKcpSession/2{NODE_ID}", nodeId)

	f.kcpScLock.Lock()
	swchan := make(chan *ServerWrite)
	existing := f.KcpServerConnections[nodeId]
	f.KcpServerConnections[nodeId] = &ServerHandle{
		conn:      conn,
		writeChan: swchan,
	}
	f.kcpScLock.Unlock()

	if existing != nil && existing.conn != nil {
		existing.conn.Close()
	}

	// Handle the KCP connection just like a WebSocket connection
	go f.handleServerConnection(nodeId, swchan, conn, func() {
		f.kcpScLock.Lock()
		delete(f.KcpServerConnections, nodeId)
		f.kcpScLock.Unlock()
	})
}
