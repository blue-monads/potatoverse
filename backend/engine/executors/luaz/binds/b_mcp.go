package binds

import (
	"context"

	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	lua "github.com/yuin/gopher-lua"
)

/*

create_client
	- list_tools
	- list_resources
	- call_tool

*/

func BindMCP(L *lua.LState) int {

	table := L.NewTable()

	createHttpClient := func(L *lua.LState) int {
		endpoint := L.CheckString(1)
		name := L.CheckString(2)
		transportType := L.OptString(3, "http")

		client := mcp.NewClient(&mcp.Implementation{Name: name, Version: "v1.0.0"}, nil)
		var transport mcp.Transport

		if transportType == "sse" {
			transport = &mcp.SSEClientTransport{
				Endpoint: endpoint,
			}
		} else {
			transport = &mcp.StreamableClientTransport{
				Endpoint: endpoint,
			}
		}

		session, err := client.Connect(context.Background(), transport, nil)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		subTable := L.NewTable()

		listTools := func(L *lua.LState) int {
			params := mcp.ListToolsParams{}
			err := luaplus.MapToStruct(L, L.CheckTable(2), &params)
			if err != nil {
				return pushError(L, err)
			}

			tools, err := session.ListTools(context.Background(), &params)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, tool := range tools.Tools {
				toolTable, err := luaplus.StructToTable(L, tool)
				if err != nil {
					return pushError(L, err)
				}

				table.Append(toolTable)
			}
			L.Push(table)
			return 1
		}
		listResources := func(L *lua.LState) int {
			params := mcp.ListResourcesParams{}

			err := luaplus.MapToStruct(L, L.CheckTable(2), &params)
			if err != nil {
				return pushError(L, err)
			}

			resources, err := session.ListResources(context.Background(), &params)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, resource := range resources.Resources {
				resourceTable, err := luaplus.StructToTable(L, resource)
				if err != nil {
					return pushError(L, err)
				}
				table.Append(resourceTable)
			}
			L.Push(table)
			return 1
		}
		callTool := func(L *lua.LState) int {
			params := mcp.CallToolParams{}

			err := luaplus.MapToStruct(L, L.CheckTable(1), &params)
			if err != nil {
				return pushError(L, err)
			}

			result, err := session.CallTool(context.Background(), &params)
			if err != nil {
				return pushError(L, err)
			}
			resultTable, err := luaplus.StructToTable(L, result)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(resultTable)
			return 1
		}

		L.SetFuncs(subTable, map[string]lua.LGFunction{
			"list_tools":     listTools,
			"list_resources": listResources,
			"call_tool":      callTool,
		})

		L.Push(subTable)

		return 1
	}

	L.SetFuncs(table, map[string]lua.LGFunction{
		"create_http_client": createHttpClient,
	})

	L.Push(table)

	return 1

}

/*

curl -X POST https://echo.mcp.inevitable.fyi/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -H "Accept: text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {
        "roots": {
          "listChanged": true
        },
        "sampling": {}
      },
      "clientInfo": {
        "name": "ExampleClient",
        "version": "1.0.0"
      }
    }
  }'
event: message
data: {
	"result":{
		"protocolVersion":"2024-11-05",
		"capabilities":{
			"resources":{"listChanged":true},
			"tools":{"listChanged":true},
			"prompts":{"listChanged":true}
		},
		"serverInfo":{
			"name":"Echo",
			"version":"1.0.0"
			}
		},
		"jsonrpc":"2.0",
		"id":1
	}
}

*/
