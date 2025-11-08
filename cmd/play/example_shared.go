package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Example demonstrating how to use SharedServerTransport
func runSharedServerExample() {
	ctx := context.Background()

	// Step 1: Create a single server instance
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "shared-hub",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		HasTools:     true,
		HasResources: true,
	})

	// Step 2: Configure the server with tools and resources
	type EchoArgs struct {
		Message string `json:"message" jsonschema:"Message to echo"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echoes a message",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args EchoArgs) (*mcp.CallToolResult, struct{}, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Echo: " + args.Message},
			},
		}, struct{}{}, nil
	})

	server.AddResource(&mcp.Resource{
		Name:        "shared-data",
		Description: "Shared resource accessible to all clients",
		URI:         "shared://data",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					Text: "This is shared data that all clients can access",
					URI:  "shared://data",
				},
			},
		}, nil
	})

	// Step 3: Create transport with the shared server
	clientTransport, serverTransport := NewSharedServerTransport()

	go func() {
		if err := server.Run(ctx, serverTransport); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	log.Println("Shared server started, spawning multiple clients...")

	// Step 5: Create multiple clients that share the same server
	const numClients = 30
	var wg sync.WaitGroup

	for i := range numClients {
		wg.Add(1)
		clientID := i + 1

		go func(id int) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			clientName := fmt.Sprintf("client-%d", id)
			log.Printf("[%s] Connecting to shared server...\n", clientName)

			// Create client and connect to the shared server
			client := mcp.NewClient(&mcp.Implementation{
				Name:    clientName,
				Version: "1.0.0",
			}, nil)

			log.Printf("[%s] Created client\n", clientName)

			session, err := client.Connect(ctx, clientTransport, nil)
			if err != nil {
				log.Printf("[%s] Connect error: %v\n", clientName, err)
				return
			}

			log.Printf("[%s] Connected to shared server\n", clientName)

			defer session.Close()

			log.Printf("[%s] Connected successfully\n", clientName)

			// Call the echo tool
			result, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name: "echo",
				Arguments: map[string]interface{}{
					"message": fmt.Sprintf("Hello from %s", clientName),
				},
			})
			if err != nil {
				log.Printf("[%s] CallTool error: %v\n", clientName, err)
			} else {

				innertext := result.Content[0].(*mcp.TextContent).Text
				fullClientName := fmt.Sprintf("client-%d", id)

				if !strings.Contains(innertext, fullClientName) {
					panic(fmt.Sprintf("expected %s to contain %s", innertext, fullClientName))
				}

				qq.Println("result", fullClientName, "|>", innertext)
				log.Printf("[%s] Tool result: %v\n", clientName, result.Content)
			}

			// Read the shared resource
			resource, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
				URI: "shared://data",
			})
			if err != nil {
				log.Printf("[%s] ReadResource error: %v\n", clientName, err)
			} else {
				log.Printf("[%s] Resource content: %s\n", clientName, resource.Contents[0].Text)
			}

			log.Printf("[%s] Completed\n", clientName)
		}(clientID)

		// Stagger client starts slightly
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for all clients to finish
	log.Println("Waiting for all clients to complete...")
	wg.Wait()

	log.Println("\n=== Summary ===")
	log.Printf("All %d clients successfully shared a single server instance\n", numClients)
	log.Println("Benefits: Lower memory usage, shared state, single point of configuration")

	// Step 6: Cleanup - shutdown the shared server
	if err := clientTransport.Shutdown(); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	}

	log.Println("Shared server example completed!")
}
