package main

import (
	"bytes"
	"context"
	"log"

	"github.com/blue-monads/turnix/backend/labs/mcphub"
	"github.com/k0kubun/pp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func runMcp() {

	reader := bytes.NewReader(make([]byte, 1024))
	writer := bytes.NewBuffer(make([]byte, 1024))

	server := mcp.NewServer(&mcp.Implementation{Name: "mcphub"}, &mcp.ServerOptions{
		HasTools: true,
	})

	// server.AddTool(&mcp.Tool{
	// 	Name:        "easyws",
	// 	Description: "Easy WebSocket",
	// }, EasyWsHandler)

	server.AddResource(&mcp.Resource{
		Name:        "easyws",
		Description: "Easy WebSocket",
	}, EasyWsResourceHandler)

	err := server.Run(context.Background(), mcphub.NewPotatoTransport(reader, writer))
	if err != nil {
		log.Fatal(err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "mcphub"}, nil)
	session, err := client.Connect(context.Background(), mcphub.NewPotatoTransport(reader, writer), nil)
	if err != nil {
		log.Fatal(err)
	}

	r, err := session.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		log.Fatal(err)
	}

	pp.Println("@tools", r.Tools)

	defer session.Close()

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
