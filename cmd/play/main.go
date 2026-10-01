package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

const addr = ":8080"

// mqttConnectByte is the fixed header of an MQTT CONNECT packet, which is
// always the first packet a client sends.
const mqttConnectByte = 0x10

func main() {
	root, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	httpL := newChanListener(root.Addr())
	mqttL := newChanListener(root.Addr())

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, World!")
		})
		log.Fatal(http.Serve(httpL, mux))
	}()

	broker := mqtt.New(&mqtt.Options{InlineClient: true})
	if err := broker.AddHook(new(auth.AllowHook), nil); err != nil {
		log.Fatal(err)
	}
	if err := broker.AddListener(listeners.NewNet("mux", mqttL)); err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := broker.Serve(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("http + mqtt listening on %s", addr)

	for {
		conn, err := root.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go route(conn, httpL, mqttL)
	}
}

func route(conn net.Conn, httpL, mqttL *chanListener) {
	br := bufio.NewReader(conn)
	first, err := br.Peek(1)
	if err != nil {
		conn.Close()
		return
	}

	pc := &peekedConn{Conn: conn, r: br}
	if first[0] == mqttConnectByte {
		mqttL.push(pc)
	} else {
		httpL.push(pc)
	}
}

type peekedConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *peekedConn) Read(p []byte) (int, error) { return c.r.Read(p) }

type chanListener struct {
	addr   net.Addr
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newChanListener(addr net.Addr) *chanListener {
	return &chanListener{
		addr:   addr,
		conns:  make(chan net.Conn),
		closed: make(chan struct{}),
	}
}

func (l *chanListener) push(c net.Conn) {
	select {
	case l.conns <- c:
	case <-l.closed:
		c.Close()
	}
}

func (l *chanListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.conns:
		return c, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *chanListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *chanListener) Addr() net.Addr { return l.addr }
