package mcphub

import (
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

/*

MCPHub (server)

- easyws
- kvstore

*/

type MCPhub struct {
	server *mcp.Server
	app    xtypes.App
}
