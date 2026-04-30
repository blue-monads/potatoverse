package xwebsocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability/emtyctx"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var (
	Name = "xWebsocket"
	Icon = `<i class="fa-solid fa-plug"></i>`

	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &WebsocketBuilder{
				app:    appTyped,
				signer: appTyped.Signer(),
				engine: appTyped.Engine().(xtypes.Engine),
			}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type WebsocketBuilder struct {
	app    xtypes.App
	signer *signer.Signer
	engine xtypes.Engine
}

func (b *WebsocketBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()
	spaceId, err := handle.GetSpaceId()
	if err != nil {
		return nil, err
	}

	return &WebsocketCapability{
		builder:      b,
		spaceId:      spaceId,
		installId:    model.InstallID,
		capabilityId: model.ID,
		connections:  make(map[string]*wsConn),
	}, nil
}

func (b *WebsocketBuilder) Serve(ctx *gin.Context) {}

func (b *WebsocketBuilder) Name() string {
	return Name
}

func (b *WebsocketBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}

type wsMsg struct {
	data   []byte
	binary bool
}

// wsConn represents a managed websocket connection
type wsConn struct {
	connId string
	userId int64
	conn   net.Conn
	send   chan *wsMsg
	once   sync.Once
	closed bool
}

func (wc *wsConn) teardown() {
	wc.once.Do(func() {
		wc.closed = true
		wc.send <- nil
		wc.conn.Close()
	})
}

type WebsocketCapability struct {
	builder      *WebsocketBuilder
	installId    int64
	capabilityId int64
	spaceId      int64

	connections map[string]*wsConn
	mu          sync.RWMutex
}

var ErrInvalidToken = errors.New("invalid token")

// Handle validates the token and emits handle_websocket_connect.
// The upgrade only happens when the action handler calls finish_upgrade.
func (c *WebsocketCapability) Handle(ctx *gin.Context) {
	token := ctx.Request.URL.Query().Get("token")
	if token == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claim, err := c.parseToken(token)
	if err != nil {
		qq.Println("@ws/invalid_token", err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cctx := &connectCtx{
		cap:    c,
		ginCtx: ctx,
		claim:  claim,
	}

	err = c.builder.engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    c.spaceId,
		EventType:  "capability",
		ActionName: "on_websocket_connect",
		Params: map[string]string{
			"conn_id":       claim.ResourceId,
			"capability_id": fmt.Sprintf("%d", c.capabilityId),
			"capability":    "websocket",
			"user_id":       fmt.Sprintf("%d", claim.UserId),
		},
		Request: cctx,
	})

	if err != nil {
		qq.Println("@ws/connect_event_error", err)
	}

	if cctx.wc == nil {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	go c.writePump(cctx.wc)
	go c.readPump(cctx.wc)
}

func (c *WebsocketCapability) parseToken(token string) (*signer.CapabilityClaim, error) {
	claim, err := c.builder.signer.ParseCapability(token)
	if err != nil {
		return nil, err
	}

	qq.Println("@ws/claim", claim)

	if claim.InstallId != c.installId {
		qq.Println("@wrong_install_id", claim.InstallId, c.installId)
		return nil, ErrInvalidToken
	}

	if claim.CapabilityId != c.capabilityId {
		qq.Println("@wrong_capability_id", claim.CapabilityId, c.capabilityId)
		return nil, ErrInvalidToken
	}

	return claim, nil
}

func (c *WebsocketCapability) writePump(wc *wsConn) {
	defer func() {
		wc.conn.Close()
		if !wc.closed {
			c.removeConn(wc.connId)
		}
	}()

	for msg := range wc.send {
		if msg == nil {
			return
		}

		var err error
		if msg.binary {
			err = wsutil.WriteServerBinary(wc.conn, msg.data)
		} else {
			err = wsutil.WriteServerText(wc.conn, msg.data)
		}
		if err != nil {
			qq.Println("@ws/write_error", wc.connId, err)
			c.removeConn(wc.connId)
			return
		}

		if wc.closed {
			return
		}
	}
}

func (c *WebsocketCapability) readPump(wc *wsConn) {
	for {
		if wc.closed {
			return
		}

		data, opCode, err := wsutil.ReadClientData(wc.conn)
		if err != nil {
			if !wc.closed {
				c.removeConn(wc.connId)
			}
			return
		}

		switch opCode {
		case ws.OpClose:
			c.removeConn(wc.connId)
			return
		case ws.OpPing:
			wsutil.WriteServerMessage(wc.conn, ws.OpPong, nil)
		case ws.OpPong:
			continue
		case ws.OpText, ws.OpBinary:
			c.handleMessage(wc, data)
		}
	}
}

func (c *WebsocketCapability) handleMessage(wc *wsConn, data []byte) {
	ctx := &messageCtx{
		cap:     c,
		payload: data,
	}

	err := c.builder.engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    c.spaceId,
		EventType:  "capability",
		ActionName: "on_websocket_message",
		Params: map[string]string{
			"conn_id":       wc.connId,
			"capability_id": fmt.Sprintf("%d", c.capabilityId),
			"capability":    "websocket",
			"user_id":       fmt.Sprintf("%d", wc.userId),
		},
		Request: ctx,
	})

	if err != nil {
		qq.Println("@ws/handle_message_error", wc.connId, err)
	}
}

func (c *WebsocketCapability) removeConn(connId string) {
	c.mu.Lock()
	wc, exists := c.connections[connId]
	if exists {
		delete(c.connections, connId)
	}
	c.mu.Unlock()

	if exists && wc != nil {
		wc.teardown()

		_ = c.builder.engine.EmitActionEvent(&xtypes.ActionEventOptions{
			SpaceId:    c.spaceId,
			EventType:  "capability",
			ActionName: "on_websocket_disconnect",
			Params: map[string]string{
				"conn_id":       connId,
				"capability_id": fmt.Sprintf("%d", c.capabilityId),
				"capability":    "websocket",
				"user_id":       fmt.Sprintf("%d", wc.userId),
			},
			Request: emtyctx.Instance,
		})
	}
}

// shared helpers

func (c *WebsocketCapability) sendToConnections(connIds []string, message []byte, binary bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	qq.Println("@ws/send_to_connections", len(connIds))

	msg := &wsMsg{data: message, binary: binary}

	for _, connId := range connIds {

		qq.Println("@ws/send_message", connId)

		wc, exists := c.connections[connId]
		if !exists || wc.closed {
			continue
		}

		select {
		case wc.send <- msg:
		default:
			qq.Println("@ws/drop_message", connId)
		}
	}

	return nil
}

func (c *WebsocketCapability) broadcastAll(message []byte, binary bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	msg := &wsMsg{data: message, binary: binary}

	for connId, wc := range c.connections {
		if wc.closed {
			continue
		}

		select {
		case wc.send <- msg:
		default:
			qq.Println("@ws/drop_message", connId)
		}
	}

	return nil
}

func (c *WebsocketCapability) listConnections() []map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]map[string]any, 0, len(c.connections))
	for _, wc := range c.connections {
		if wc.closed {
			continue
		}

		result = append(result, map[string]any{
			"conn_id": wc.connId,
			"user_id": wc.userId,
		})
	}

	return result
}

// Capability interface

func (c *WebsocketCapability) ListActions() ([]string, error) {
	return []string{
		"send_to_connections",
		"broadcast_message",
		"list_connections",
	}, nil
}

func (c *WebsocketCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	qq.Println("@ws/execute", name)

	switch name {
	case "send_to_connections":
		var p struct {
			ConnIds []string `json:"conns"`
			Message any      `json:"message"`
			Binary  bool     `json:"binary"`
		}
		if err := params.AsJson(&p); err != nil {
			qq.Println("@ws/send_to_connections asjson error", err)
			return nil, err
		}

		qq.Println("@ws/send_to_connections params", p.ConnIds)

		anyBytes, err := json.Marshal(p.Message)
		if err != nil {
			qq.Println("@ws/send_to_connections marshal error", err)
			return nil, err
		}

		if len(p.ConnIds) == 0 {
			qq.Println("@ws/send_to_connections conns is required")
			return nil, errors.New("conns is required")
		}
		return ok, c.sendToConnections(p.ConnIds, anyBytes, p.Binary)

	case "broadcast_message":
		var p struct {
			Message json.RawMessage `json:"message"`
			Binary  bool            `json:"binary"`
		}
		if err := params.AsJson(&p); err != nil {
			return nil, err
		}
		return ok, c.broadcastAll(p.Message, p.Binary)

	case "list_connections":
		return c.listConnections(), nil

	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *WebsocketCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return c, nil
}

func (c *WebsocketCapability) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, wc := range c.connections {
		wc.teardown()
	}
	c.connections = make(map[string]*wsConn)

	return nil
}

var ok = map[string]any{"success": true}
