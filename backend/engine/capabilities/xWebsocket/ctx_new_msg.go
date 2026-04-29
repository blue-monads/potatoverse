package xwebsocket

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability/easyaction"
)

// messageCtx is the ActionRequest for handle_websocket_message.
// Provides easyaction methods on the payload plus websocket-specific actions.
type messageCtx struct {
	cap     *WebsocketCapability
	payload []byte
}

func (m *messageCtx) ListActions() ([]string, error) {
	return append(easyaction.Methods,
		"broadcast_current_message",
		"broadcast_message",
		"send_to_connections",
		"list_connections",
	), nil
}

func (m *messageCtx) ExecuteAction(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "broadcast_current_message":
		var p struct {
			ConnIds []string `json:"conns"`
			Binary  bool     `json:"binary"`
		}
		if err := params.AsJson(&p); err != nil {
			return nil, err
		}

		if len(p.ConnIds) == 0 {
			return ok, m.cap.broadcastAll(m.payload, p.Binary)
		}
		return ok, m.cap.sendToConnections(p.ConnIds, m.payload, p.Binary)

	case "broadcast_message", "send_to_connections", "list_connections":
		return m.cap.Execute(name, params)

	default:
		resp, err := easyaction.BytelazyDataActions(m.payload, name, params)
		if err != nil {
			if errors.Is(err, easyaction.ErrUnknownAction) {
				return nil, errors.New("unknown action: " + name)
			}
			return nil, err
		}
		return resp, nil
	}
}
