package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SharedServerTransport allows multiple clients to share a single MCP server instance.
// It multiplexes messages from multiple client connections to a single server.
type SharedServerTransport struct {
	// The shared server instance
	// server *mcp.Server

	// Active client connections
	clients map[string]*clientConnection
	mu      sync.RWMutex

	// Aggregated channel that combines messages from all clients
	aggregatedCh chan messageWithClient

	// Request ID to client mapping for routing responses
	requestMap map[interface{}]string
	requestMu  sync.RWMutex

	// Multiplexed connection for the server
	serverConn *multiplexedConnection

	// Used to generate unique connection IDs
	nextID atomic.Int64
}

// messageWithClient pairs a message with the client that sent it
type messageWithClient struct {
	msg      jsonrpc.Message
	clientID string
}

// clientConnection represents a single client's connection
type clientConnection struct {
	id        string
	inCh      chan jsonrpc.Message // Messages from server to this client
	outCh     chan jsonrpc.Message // Messages from this client to server
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// multiplexedConnection aggregates all client connections into a single connection
// that the server interacts with
type multiplexedConnection struct {
	transport *SharedServerTransport
	connID    string
	closed    atomic.Bool
}

// NewSharedServerTransport creates a transport where all clients share a single server.
func NewSharedServerTransport(server *mcp.Server) *SharedServerTransport {
	t := &SharedServerTransport{
		//	server:       server,
		clients:      make(map[string]*clientConnection),
		aggregatedCh: make(chan messageWithClient, 100),
		requestMap:   make(map[interface{}]string),
	}

	// Create the multiplexed connection for the server
	t.serverConn = &multiplexedConnection{
		transport: t,
		connID:    "shared-server",
	}

	return t
}

// Connect implements mcp.Transport for clients. Each client gets its own connection,
// but all share the same underlying server.
func (t *SharedServerTransport) Connect(ctx context.Context) (mcp.Connection, error) {

	clientID := fmt.Sprintf("client-%d", t.nextID.Add(1))

	// Create client connection
	client := &clientConnection{
		id:    clientID,
		inCh:  make(chan jsonrpc.Message, 100),
		outCh: make(chan jsonrpc.Message, 100),
	}
	client.ctx, client.cancel = context.WithCancel(ctx)

	// Register client
	t.mu.Lock()
	t.clients[clientID] = client
	t.mu.Unlock()

	// Start goroutine to forward messages from client to server
	go t.forwardClientToServer(client)

	// Return client-side connection
	return &channelConn{
		connID: clientID,
		inCh:   client.inCh,
		outCh:  client.outCh,
		closeFunc: func() {
			client.closeOnce.Do(func() {
				client.cancel()
				t.mu.Lock()
				delete(t.clients, clientID)
				t.mu.Unlock()

				// Clean up request mappings for this client
				t.requestMu.Lock()
				for reqID, cID := range t.requestMap {
					if cID == clientID {
						delete(t.requestMap, reqID)
					}
				}
				t.requestMu.Unlock()

				close(client.outCh)
				close(client.inCh)
			})
		},
	}, nil
}

// forwardClientToServer reads from client and forwards to the aggregated channel
func (t *SharedServerTransport) forwardClientToServer(client *clientConnection) {
	for {
		select {
		case msg, ok := <-client.outCh:
			if !ok {
				return
			}
			// Track request ID for response routing
			t.trackRequestFromClient(msg, client.id)

			// Forward to aggregated channel
			select {
			case t.aggregatedCh <- messageWithClient{msg: msg, clientID: client.id}:
			case <-client.ctx.Done():
				return
			}
		case <-client.ctx.Done():
			return
		}
	}
}

// trackRequestFromClient extracts request ID and maps it to the client
func (t *SharedServerTransport) trackRequestFromClient(msg jsonrpc.Message, clientID string) {
	// Try to extract ID from the message
	// JSON-RPC messages have an "id" field for requests
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	var msgMap map[string]interface{}
	if err := json.Unmarshal(data, &msgMap); err != nil {
		return
	}

	if id, ok := msgMap["id"]; ok && id != nil {
		t.requestMu.Lock()
		t.requestMap[id] = clientID
		t.requestMu.Unlock()
	}
}

// getClientForResponse extracts response ID and finds the corresponding client
func (t *SharedServerTransport) getClientForResponse(msg jsonrpc.Message) (string, bool) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", false
	}

	var msgMap map[string]interface{}
	if err := json.Unmarshal(data, &msgMap); err != nil {
		return "", false
	}

	if id, ok := msgMap["id"]; ok && id != nil {
		t.requestMu.RLock()
		clientID, found := t.requestMap[id]
		t.requestMu.RUnlock()

		// Clean up the mapping after use
		if found {
			t.requestMu.Lock()
			delete(t.requestMap, id)
			t.requestMu.Unlock()
		}

		return clientID, found
	}

	return "", false
}

// Shutdown stops the shared server and closes all client connections
func (t *SharedServerTransport) Shutdown() error {

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, client := range t.clients {
		client.cancel()
		close(client.outCh)
		close(client.inCh)
	}
	t.clients = make(map[string]*clientConnection)

	close(t.aggregatedCh)

	return nil
}

// multiplexedConnection.Read aggregates reads from all clients
func (m *multiplexedConnection) Read(ctx context.Context) (jsonrpc.Message, error) {
	if m.closed.Load() {
		return nil, errors.New("multiplexed connection closed")
	}

	// Read from the aggregated channel that combines all client messages
	select {
	case msgWithClient, ok := <-m.transport.aggregatedCh:
		if !ok {
			return nil, errors.New("aggregated channel closed")
		}
		return msgWithClient.msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// multiplexedConnection.Write routes response to the appropriate client
func (m *multiplexedConnection) Write(ctx context.Context, msg jsonrpc.Message) error {
	if m.closed.Load() {
		return errors.New("multiplexed connection closed")
	}

	// Try to find the specific client for this response
	if clientID, found := m.transport.getClientForResponse(msg); found {
		m.transport.mu.RLock()
		client, exists := m.transport.clients[clientID]
		m.transport.mu.RUnlock()

		if exists {
			select {
			case client.inCh <- msg:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			default:
				return errors.New("client channel full")
			}
		}
	}

	// If we can't find a specific client, it might be a notification
	// Broadcast to all clients (useful for server-initiated messages)
	m.transport.mu.RLock()
	defer m.transport.mu.RUnlock()

	for _, client := range m.transport.clients {
		select {
		case client.inCh <- msg:
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Skip clients with full channels
		}
	}

	return nil
}

func (m *multiplexedConnection) Close() error {
	m.closed.Store(true)
	return nil
}

func (m *multiplexedConnection) SessionID() string {
	return m.connID
}

// sharedServerTransportWrapper wraps SharedServerTransport to provide the server
// with its multiplexed connection
type sharedServerTransportWrapper struct {
	transport *SharedServerTransport
	once      sync.Once
}

func (w *sharedServerTransportWrapper) Connect(ctx context.Context) (mcp.Connection, error) {
	// Always return the same multiplexed connection
	return w.transport.serverConn, nil
}
