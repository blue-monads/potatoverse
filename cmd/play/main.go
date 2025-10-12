package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	runMcp()
}

func Main1() {
	ctx := context.Background()

	// Create a new client, with no features.
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)

	// Connect to a server over stdin/stdout.
	transport := &mcp.SSEServerTransport{
		Endpoint: "http://localhost:8080/mcp",
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Call a tool on the server.
	params := &mcp.CallToolParams{
		Name:      "greet",
		Arguments: map[string]any{"name": "you"},
	}

	res, err := session.CallTool(ctx, params)
	if err != nil {
		log.Fatalf("CallTool failed: %v", err)
	}
	if res.IsError {
		log.Fatal("tool failed")
	}
	for _, c := range res.Content {
		log.Print(c.(*mcp.TextContent).Text)
	}
}

func Main2() {
	fmt.Println("Hello, World!")

	fistPart := "potato_asbasu66612jhagshasg___&aksa"
	secondPart := "verse_asbasu66612jhagshasg___&aksa"

	// calculate sha1 hash of "potato"
	hash := sha1.New()
	hash.Write([]byte(fistPart + secondPart))
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))

	// calculate sha1 hash of "potatoverse"
	hash = sha1.New()
	hash.Write([]byte(fistPart))
	hash.Write([]byte(secondPart))
	fmt.Println(hex.EncodeToString(hash.Sum(nil)))

}
