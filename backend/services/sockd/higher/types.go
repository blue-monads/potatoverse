package higher

import (
	"encoding/json"
)

// Message represents a websocket message
type Message struct {
	Id     int64           `json:"id"`
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
	Topic  string          `json:"topic,omitempty"`
	Target int64           `json:"target,omitempty"`
}

// publishEvent is used internally for room event loop
type publishEvent struct {
	topic   string
	message []byte
}

// directMessageEvent is used internally for room event loop
type directMessageEvent struct {
	targetConnId int64
	message      []byte
}
