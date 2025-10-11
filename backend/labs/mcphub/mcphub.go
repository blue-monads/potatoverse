package mcphub

import "github.com/blue-monads/turnix/backend/xtypes"

type MCPhub struct {
}

type McpProvider struct {
	Name        string
	Description string
	Tools       []McpTool
	Resources   []McpResource
	Handle      func(handlerName string, params any) (any, error)
}

type McpTool struct {
	Name        string
	Description string
}

type McpResource struct {
	Name        string
	Description string
}

type MCPBuilder func(app xtypes.App) (McpProvider, error)
