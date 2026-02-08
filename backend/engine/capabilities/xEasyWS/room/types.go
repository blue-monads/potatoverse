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
	Type       string `json:"type"`
	SubType    string `json:"sub_type,omitempty"`
	FromConnId string `json:"from_cid,omitempty"`
	ToConnId   string `json:"to_cid,omitempty"`
	RequestId  string `json:"rid,omitempty"`
	Topic      string `json:"topic,omitempty"`
	Target     string `json:"target,omitempty"` // service_target (eg terminal_12)

	Data json.RawMessage `json:"data"`
}

// publishEvent is used internally for room event loop
type publishEvent struct {
	topic   string
	message []byte
	connId  ConnId
}

// directMessageEvent is used internally for room event loop
type directMessageEvent struct {
	targetConnId ConnId
	message      []byte
}
