package room

import (
	"encoding/json"
)

const (
	MessageTypeBroadcast           = "sbroadcast"
	MessageTypePublish             = "spublish"
	MessageTypeDirectMessage       = "sdirect"
	ClientMessageTypeBroadcast     = "cbroadcast"
	ClientMessageTypePublish       = "cpublish"
	ClientMessageTypeDirectMessage = "cdirect"
	ClientMessageTypeGetPresence   = "cpresence"
	ClientMessageTypeCommand       = "ccommand"
)

// Message represents a websocket message
type Message struct {
	Type   string          `json:"type"`
	Data   json.RawMessage `json:"data"`
	Topic  string          `json:"topic,omitempty"` // or command
	Target string          `json:"target,omitempty"`
}

// publishEvent is used internally for room event loop
type publishEvent struct {
	topic   string
	message []byte
}

// directMessageEvent is used internally for room event loop
type directMessageEvent struct {
	targetConnId ConnId
	message      []byte
}
