package xwebsocket

import (
	"errors"
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

// connectCtx is the ActionRequest for handle_websocket_connect.
// The handler must call finish_upgrade to accept and upgrade the connection.
type connectCtx struct {
	cap    *WebsocketCapability
	ginCtx *gin.Context
	claim  *signer.CapabilityClaim
	wc     *wsConn
}

func (cc *connectCtx) ListActions() ([]string, error) {
	return []string{"finish_upgrade", "list_connections"}, nil
}

func (cc *connectCtx) ExecuteAction(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "finish_upgrade":
		if cc.wc != nil {
			return map[string]any{"success": true, "conn_id": cc.wc.connId}, nil
		}

		conn, _, _, err := ws.UpgradeHTTP(cc.ginCtx.Request, cc.ginCtx.Writer)
		if err != nil {
			return nil, fmt.Errorf("failed to upgrade: %w", err)
		}

		wc := &wsConn{
			connId: cc.claim.ResourceId,
			userId: cc.claim.UserId,
			conn:   conn,
			send:   make(chan *wsMsg, 16),
		}

		cc.cap.mu.Lock()
		existing := cc.cap.connections[wc.connId]
		cc.cap.connections[wc.connId] = wc
		cc.cap.mu.Unlock()

		if existing != nil {
			existing.teardown()
		}

		cc.wc = wc
		return map[string]any{"success": true, "conn_id": wc.connId}, nil

	case "list_connections":
		return cc.cap.listConnections(), nil

	default:
		return nil, errors.New("unknown action: " + name)
	}
}
