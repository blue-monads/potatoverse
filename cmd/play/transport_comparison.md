# MCP Transport Comparison: Per-Client Server vs Shared Server

## Approach 1: Per-Client Server (Current Implementation in `transport.go`)

### How it Works
- Each client connection spawns its **own dedicated server instance**
- Each server is independent with its own state
- No message routing complexity

### Pros
- ✅ **Simple implementation** - no need for complex message multiplexing
- ✅ **Isolated state** - each client has completely isolated server state
- ✅ **No routing issues** - messages naturally flow between client and its server
- ✅ **Better for stateful servers** - each client can have its own session state
- ✅ **Easier to debug** - clear 1:1 relationship between client and server

### Cons
- ❌ **Higher memory usage** - N clients = N server instances
- ❌ **No shared state** - clients can't see each other's data
- ❌ **More resource intensive** - each server has its own goroutines and resources

### Use Cases
- When each client needs isolated state
- When server instances are lightweight
- When you don't need inter-client communication
- Most typical MCP scenarios

### Code Example
```go
// Create server factory
serverFactory := func() *mcp.Server {
    server := mcp.NewServer(&mcp.Implementation{Name: "myserver", Version: "1.0"}, nil)
    // Configure server...
    return server
}

// Create transport
transport := NewChannelTransport(100, serverFactory)

// Each client.Connect() creates a new server instance
client1 := mcp.NewClient(&mcp.Implementation{Name: "client1", Version: "1.0"}, nil)
session1, _ := client1.Connect(ctx, transport, nil) // Server instance 1

client2 := mcp.NewClient(&mcp.Implementation{Name: "client2", Version: "1.0"}, nil)
session2, _ := client2.Connect(ctx, transport, nil) // Server instance 2 (separate!)
```

---

## Approach 2: Shared Server (New Implementation in `shared_transport.go`)

### How it Works
- **Single server instance** serves all clients
- Message multiplexing routes requests from clients to server
- Response routing based on JSON-RPC request IDs
- Aggregated channel fan-in pattern

### Pros
- ✅ **Lower memory footprint** - one server instance regardless of client count
- ✅ **Shared state possible** - all clients interact with same server state
- ✅ **Resource efficient** - single set of server goroutines
- ✅ **Better for collaborative scenarios** - clients can share data through server

### Cons
- ❌ **Complex implementation** - requires message routing and multiplexing
- ❌ **Potential bottleneck** - single server handling all client requests
- ❌ **Concurrency concerns** - server must be thread-safe for concurrent clients
- ❌ **Harder to debug** - message routing can obscure issues
- ❌ **Request ID tracking overhead** - need to map requests to clients

### Use Cases
- When server instances are resource-heavy
- When you need shared state across clients
- When implementing collaborative features
- When running many clients (100+)
- Hub/room patterns (e.g., chat server, game server)

### Code Example
```go
// Create single server instance
server := mcp.NewServer(&mcp.Implementation{Name: "myserver", Version: "1.0"}, &mcp.ServerOptions{
    HasTools: true,
    HasResources: true,
})

// Configure server (tools, resources, etc.)
mcp.AddTool(server, &mcp.Tool{Name: "test"}, handler)

// Create shared transport
transport := NewSharedServerTransport(server)

// Start the server ONCE
transport.StartServer()

// Multiple clients share the same server
client1 := mcp.NewClient(&mcp.Implementation{Name: "client1", Version: "1.0"}, nil)
session1, _ := client1.Connect(ctx, transport, nil) // Uses shared server

client2 := mcp.NewClient(&mcp.Implementation{Name: "client2", Version: "1.0"}, nil)
session2, _ := client2.Connect(ctx, transport, nil) // Uses SAME shared server

// Cleanup
transport.Shutdown()
```

---

## Architecture Diagrams

### Per-Client Server Architecture
```
Client 1 ──> [Transport] ──> Server Instance 1
                                  ↓
                            [Tools, Resources]

Client 2 ──> [Transport] ──> Server Instance 2
                                  ↓
                            [Tools, Resources]

Client 3 ──> [Transport] ──> Server Instance 3
                                  ↓
                            [Tools, Resources]
```

### Shared Server Architecture
```
                        ┌─────────────────────┐
Client 1 ──┐            │   Shared Server     │
           │            │                     │
Client 2 ──┼─> [Mux] ──>│  [Tools, Resources] │
           │            │                     │
Client 3 ──┘            │  (Single Instance)  │
                        └─────────────────────┘
         
         Mux = Multiplexed Connection
         - Aggregates client messages
         - Routes responses by request ID
```

---

## Key Implementation Differences

### Message Flow

**Per-Client:**
```go
Client → Client Conn → Channel Pair → Server Conn → Server Instance
Server Instance → Server Conn → Channel Pair → Client Conn → Client
```

**Shared:**
```go
Client → Client Conn → Aggregated Channel → Multiplexed Conn → Shared Server
                          ↑
                    (Fan-in from all clients)

Shared Server → Multiplexed Conn → Route by Request ID → Client Conn → Client
                                          ↓
                                  (Fan-out to specific client)
```

### Request ID Tracking (Shared Server Only)
```go
// When client sends request:
1. Extract request ID from message
2. Map: requestID -> clientID
3. Forward message to server

// When server sends response:
1. Extract request ID from response
2. Lookup: requestID -> clientID
3. Route to specific client
4. Clean up mapping
```

---

## Performance Comparison

| Metric | Per-Client Server | Shared Server |
|--------|-------------------|---------------|
| **Memory per client** | Higher (full server instance) | Lower (shared resources) |
| **Latency** | Lower (direct path) | Slightly higher (routing overhead) |
| **Throughput** | Independent per client | Shared bottleneck |
| **Scalability** | Limited by memory | Limited by server throughput |
| **State isolation** | Perfect (separate instances) | Requires careful design |

---

## Recommendation

**Use Per-Client Server (default) when:**
- 👤 Client count < 100
- 🔒 Each client needs isolated state
- 🎯 Simple 1:1 client-server relationships
- 🐛 Debugging and testing scenarios

**Use Shared Server when:**
- 🏢 Running 100+ clients
- 🤝 Clients need to share state
- 💾 Server instances are resource-heavy
- 🎮 Implementing collaborative features (chat, games, collaborative editing)

**For most MCP use cases, the Per-Client Server approach is recommended** due to its simplicity and natural state isolation.

