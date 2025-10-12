package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ChannelTransport is an in-memory transport for MCP using Go channels.
// Each connection gets its own pair of channels for bidirectional communication.
type ChannelTransport struct {
	// Callback function to create a new server instance for each client
	serverFactory func() *mcp.Server

	// Active connections
	connections map[string]*connectionPair
	mu          sync.RWMutex

	// Used to generate unique session IDs
	nextID atomic.Int64
}

// connectionPair represents a bidirectional channel connection
type connectionPair struct {
	// Client to Server channel
	c2s chan jsonrpc.Message
	// Server to Client channel
	s2c       chan jsonrpc.Message
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// NewChannelTransport creates a new channel-based transport.
func NewChannelTransport(bufferSize int, serverFactory func() *mcp.Server) *ChannelTransport {
	return &ChannelTransport{
		serverFactory: serverFactory,
		connections:   make(map[string]*connectionPair),
	}
}

// Connect implements the mcp.Transport interface.
// Each call creates a new isolated connection pair and spawns a server instance.
func (t *ChannelTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	connID := fmt.Sprintf("conn-%d", t.nextID.Add(1))

	// Create a new connection pair with dedicated channels
	pair := &connectionPair{
		c2s: make(chan jsonrpc.Message, 100),
		s2c: make(chan jsonrpc.Message, 100),
	}
	pair.ctx, pair.cancel = context.WithCancel(ctx)

	t.mu.Lock()
	t.connections[connID] = pair
	t.mu.Unlock()

	// Start a server instance for this connection
	go func() {
		server := t.serverFactory()

		// Create a dedicated transport for this server instance
		serverTransport := &singleConnectionTransport{
			c2s:    pair.c2s,
			s2c:    pair.s2c,
			connID: connID + "-server",
		}

		// Run the server with this dedicated transport
		if err := server.Run(pair.ctx, serverTransport); err != nil {
			// Server stopped - this is expected when connection closes
		}
	}()

	// Return the client-side connection
	return &channelConn{
		connID: connID,
		inCh:   pair.s2c, // Client reads from server-to-client channel
		outCh:  pair.c2s, // Client writes to client-to-server channel
		closeFunc: func() {
			// Only close channels once
			pair.closeOnce.Do(func() {
				pair.cancel()
				t.mu.Lock()
				delete(t.connections, connID)
				t.mu.Unlock()

				// Close the channels
				close(pair.c2s)
				close(pair.s2c)
			})
		},
	}, nil
}

// channelConn represents one side of an MCP connection.
type channelConn struct {
	connID    string
	inCh      chan jsonrpc.Message // Channel to READ messages FROM
	outCh     chan jsonrpc.Message // Channel to WRITE messages TO
	closeFunc func()               // Optional cleanup function
	mu        sync.Mutex           // Protects the close state
	closed    bool
}

func (c *channelConn) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case msg, ok := <-c.inCh:
		if !ok {
			return nil, errors.New("connection channel closed")
		}
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *channelConn) Write(ctx context.Context, msg jsonrpc.Message) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("connection is closed")
	}
	c.mu.Unlock()

	select {
	case c.outCh <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *channelConn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	// Call cleanup function if provided (this handles channel cleanup at transport level)
	if c.closeFunc != nil {
		c.closeFunc()
	}

	return nil
}

func (c *channelConn) SessionID() string {
	return c.connID
}

// singleConnectionTransport wraps a single channel pair into a transport
// that only ever returns one connection (used for server-side connections)
type singleConnectionTransport struct {
	c2s    chan jsonrpc.Message
	s2c    chan jsonrpc.Message
	connID string
	once   sync.Once
	conn   mcp.Connection
}

func (t *singleConnectionTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	t.once.Do(func() {
		t.conn = &channelConn{
			connID: t.connID,
			inCh:   t.c2s, // Server reads from client-to-server channel
			outCh:  t.s2c, // Server writes to server-to-client channel
		}
	})
	return t.conn, nil
}
