package eventhub

type Event struct {
	Group         string         `json:"group"`
	Name          string         `json:"name"`
	MCPResourceID string         `json:"mcp_resource_id"`
	Data          map[string]any `json:"data"`
}

type EventHub interface {
	Publish(spaceId int64, event *Event) error
}
