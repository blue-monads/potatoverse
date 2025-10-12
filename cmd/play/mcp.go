package main

import (
	"context"
	"log"
	"time"

	"github.com/k0kubun/pp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func runMcp() {
	ctx := context.Background()
	const channelBufferSize = 100

	// Create a server factory function that creates a new configured server for each client
	serverFactory := func() *mcp.Server {
		server := mcp.NewServer(&mcp.Implementation{Name: "mcphub", Version: "1.0.0"}, &mcp.ServerOptions{
			HasTools:     true,
			HasResources: true,
		})

		// Add a test tool with typed arguments
		type TestToolArgs struct {
			Message string `json:"message" jsonschema:"A test message"`
		}
		mcp.AddTool(server, &mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
		}, func(ctx context.Context, req *mcp.CallToolRequest, args TestToolArgs) (*mcp.CallToolResult, struct{}, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Tool called successfully with message: " + args.Message},
				},
			}, struct{}{}, nil
		})

		server.AddResource(&mcp.Resource{
			Name:        "easyws",
			Description: "Easy WebSocket",
			URI:         "easyws://test",
		}, EasyWsResourceHandler)

		return server
	}

	// Create transport with server factory
	transport := NewChannelTransport(channelBufferSize, serverFactory)

	log.Println("Transport ready, connecting clients...")

	// --- Client 1 Execution ---
	log.Println("Connecting client 1...")
	client1 := mcp.NewClient(&mcp.Implementation{Name: "test-client-1", Version: "1.0.0"}, nil)
	session1, err := client1.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Client 1 connect error: %v", err)
	}
	defer session1.Close()

	log.Println("Client 1 connected, listing tools...")
	r1, err := session1.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatalf("Client 1 ListTools error: %v", err)
	}

	pp.Println("@client1 tools", r1.Tools)

	// --- Second Client Execution ---
	log.Println("Connecting client 2...")
	client2 := mcp.NewClient(&mcp.Implementation{Name: "test-client-2", Version: "1.0.0"}, nil)
	session2, err := client2.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatalf("Client 2 connect error: %v", err)
	}
	defer session2.Close()

	log.Println("Client 2 connected, listing tools...")
	r2, err := session2.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Fatalf("Client 2 ListTools error: %v", err)
	}

	pp.Println("@client2 tools", r2.Tools)

	// Test calling a tool
	log.Println("Client 1 calling test tool...")
	toolResult, err := session1.CallTool(ctx, &mcp.CallToolParams{
		Name: "test-tool",
		Arguments: map[string]interface{}{
			"message": "Hello from client 1",
		},
	})
	if err != nil {
		log.Fatalf("Client 1 CallTool error: %v", err)
	}
	pp.Println("@client1 tool result", toolResult)

	time.Sleep(200 * time.Millisecond)
}

func EasyWsResourceHandler(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {

	return nil, nil
}

func EasyWsHandler(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	req.GetExtra()

	return nil, nil
}

/*

2025/10/12 18:46:14 calling "initialize": EOF
exit status 1

*/
