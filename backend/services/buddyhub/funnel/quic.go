package funnel

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/quic-go/quic-go"
)

func (f *Funnel) StartQuicServer() error {
	tlsConf, err := generateTLSConfig()
	if err != nil {
		return err
	}

	ln, err := quic.ListenAddr(":0", tlsConf, nil)
	if err != nil {
		return err
	}

	f.quicListener = ln
	f.quicPort = ln.Addr().(*net.UDPAddr).Port

	qq.Println("@Funnel/StartQuicServer/1{PORT}", f.quicPort)

	go f.handleQuicConnections()

	return nil
}

func (f *Funnel) handleQuicConnections() {
	for {
		conn, err := f.quicListener.Accept(context.Background())
		if err != nil {
			qq.Println("@Funnel/handleQuicConnections/1{ERROR}", err)
			break
		}

		go f.handleQuicSession(conn)
	}
}

func (f *Funnel) handleQuicSession(conn quic.Connection) {
	// Accept a bidirectional stream
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		qq.Println("@Funnel/handleQuicSession/1{ERROR}", err)
		return
	}

	// Initial handshake: Read token (64 bytes)
	tokenBuf := make([]byte, 64)
	_, err = io.ReadFull(stream, tokenBuf)
	if err != nil {
		qq.Println("@Funnel/handleQuicSession/2{ERROR}", err)
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

	qq.Println("@Funnel/handleQuicSession/3{NODE_ID}", nodeId)

	f.quicScLock.Lock()
	swchan := make(chan *ServerWrite)
	existing := f.QuicServerConnections[nodeId]
	f.QuicServerConnections[nodeId] = &ServerHandle{
		conn:      &quicStreamConn{Stream: stream, conn: conn},
		writeChan: swchan,
	}
	f.quicScLock.Unlock()

	if existing != nil && existing.conn != nil {
		existing.conn.Close()
	}

	// Handle the QUIC connection just like a WebSocket connection
	f.handleServerConnection(nodeId, swchan, &quicStreamConn{Stream: stream, conn: conn}, false, func() {
		f.quicScLock.Lock()
		delete(f.QuicServerConnections, nodeId)
		f.quicScLock.Unlock()
	})
}

// quicStreamConn wraps a quic.Stream to implement net.Conn
type quicStreamConn struct {
	quic.Stream
	conn quic.Connection
}

func (c *quicStreamConn) LocalAddr() net.Addr  { return c.conn.LocalAddr() }
func (c *quicStreamConn) RemoteAddr() net.Addr { return c.conn.RemoteAddr() }

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"funnel-quic"},
	}, nil
}
