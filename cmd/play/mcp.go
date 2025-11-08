package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func runMcp() {
	ctx := context.Background()
	const channelBufferSize = 100
	const numClients = 10
	const operationsPerClient = 5

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

		// Add another tool for variety
		type EchoArgs struct {
			Text  string `json:"text" jsonschema:"Text to echo"`
			Count int    `json:"count" jsonschema:"Number of times to repeat"`
		}
		mcp.AddTool(server, &mcp.Tool{
			Name:        "echo-tool",
			Description: "Echoes text multiple times",
		}, func(ctx context.Context, req *mcp.CallToolRequest, args EchoArgs) (*mcp.CallToolResult, struct{}, error) {
			result := ""
			for i := 0; i < args.Count; i++ {
				result += fmt.Sprintf("[%d] %s\n", i+1, args.Text)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: result},
				},
			}, struct{}{}, nil
		})

		server.AddResource(&mcp.Resource{
			Name:        "easyws",
			Description: "Easy WebSocket",
			URI:         "easyws://test",
		}, EasyWsResourceHandler)

		server.AddResource(&mcp.Resource{
			Name:        "test-data",
			Description: "Test data resource",
			URI:         "test://data",
		}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						Text: "This is test data content",
						URI:  "test://data",
					},
				},
			}, nil
		})

		return server
	}

	// Create transport with server factory
	transport := NewChannelTransport(channelBufferSize, serverFactory)

	log.Printf("Transport ready, spawning %d concurrent clients...\n", numClients)

	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	errorCount := 0

	// Spawn multiple clients concurrently
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		clientID := i + 1

		go func(id int) {
			defer wg.Done()

			clientName := fmt.Sprintf("test-client-%d", id)
			log.Printf("[Client %d] Starting...\n", id)

			// Create client
			client := mcp.NewClient(&mcp.Implementation{Name: clientName, Version: "1.0.0"}, nil)
			session, err := client.Connect(ctx, transport, nil)
			if err != nil {
				log.Printf("[Client %d] Connect error: %v\n", id, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}
			defer session.Close()

			log.Printf("[Client %d] Connected successfully\n", id)

			// Random delay to stagger operations
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			// Perform random operations
			for opNum := 0; opNum < operationsPerClient; opNum++ {
				operation := rand.Intn(4)

				switch operation {
				case 0:
					// List tools
					log.Printf("[Client %d] Operation %d: Listing tools...\n", id, opNum+1)
					tools, err := session.ListTools(ctx, &mcp.ListToolsParams{})
					if err != nil {
						log.Printf("[Client %d] ListTools error: %v\n", id, err)
						mu.Lock()
						errorCount++
						mu.Unlock()
					} else {
						log.Printf("[Client %d] Found %d tools\n", id, len(tools.Tools))
					}

				case 1:
					// Call test-tool
					message := fmt.Sprintf("Hello from client %d, operation %d", id, opNum+1)
					log.Printf("[Client %d] Operation %d: Calling test-tool...\n", id, opNum+1)
					result, err := session.CallTool(ctx, &mcp.CallToolParams{
						Name: "test-tool",
						Arguments: map[string]interface{}{
							"message": message,
						},
					})
					if err != nil {
						log.Printf("[Client %d] CallTool error: %v\n", id, err)
						mu.Lock()
						errorCount++
						mu.Unlock()
					} else {
						log.Printf("[Client %d] Tool result received: %d content items\n", id, len(result.Content))
					}

				case 2:
					// Call echo-tool
					log.Printf("[Client %d] Operation %d: Calling echo-tool...\n", id, opNum+1)
					_, err := session.CallTool(ctx, &mcp.CallToolParams{
						Name: "echo-tool",
						Arguments: map[string]interface{}{
							"text":  fmt.Sprintf("Client-%d-Echo", id),
							"count": rand.Intn(3) + 1,
						},
					})
					if err != nil {
						log.Printf("[Client %d] CallTool (echo) error: %v\n", id, err)
						mu.Lock()
						errorCount++
						mu.Unlock()
					} else {
						log.Printf("[Client %d] Echo result received\n", id)
					}

				case 3:
					// List resources
					log.Printf("[Client %d] Operation %d: Listing resources...\n", id, opNum+1)
					resources, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
					if err != nil {
						log.Printf("[Client %d] ListResources error: %v\n", id, err)
						mu.Lock()
						errorCount++
						mu.Unlock()
					} else {
						log.Printf("[Client %d] Found %d resources\n", id, len(resources.Resources))

						// Randomly read a resource if available
						if len(resources.Resources) > 0 && rand.Intn(2) == 0 {
							resourceURI := resources.Resources[rand.Intn(len(resources.Resources))].URI
							log.Printf("[Client %d] Reading resource: %s\n", id, resourceURI)
							_, err := session.ReadResource(ctx, &mcp.ReadResourceParams{URI: resourceURI})
							if err != nil {
								log.Printf("[Client %d] ReadResource error: %v\n", id, err)
								mu.Lock()
								errorCount++
								mu.Unlock()
							}
						}
					}
				}

				// Random delay between operations
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			}

			mu.Lock()
			successCount++
			mu.Unlock()
			log.Printf("[Client %d] Completed all operations\n", id)
		}(clientID)
	}

	// Wait for all clients to finish
	log.Println("Waiting for all clients to complete...")
	wg.Wait()

	log.Println("\n=== Test Summary ===")
	log.Printf("Total clients: %d\n", numClients)
	log.Printf("Successful completions: %d\n", successCount)
	log.Printf("Total errors encountered: %d\n", errorCount)

	// Show transport stats
	transport.mu.RLock()
	activeConnections := len(transport.connections)
	transport.mu.RUnlock()
	log.Printf("Active connections remaining: %d\n", activeConnections)

	time.Sleep(500 * time.Millisecond)
	qq.Println("Transport test completed successfully!")
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
