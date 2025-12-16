package easyws

import (
	"errors"
	"net/http"

	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

type EasyWsCapability struct {
	builder      *EasyWsBuilder
	app          xtypes.App
	spaceId      int64
	installId    int64
	capabilityId int64
	room         *room.Room

	cmdChan chan room.Message
}

func (c *EasyWsCapability) ListActions() ([]string, error) {
	return []string{
		"broadcast",
		"publish",
		"direct_message",
		"subscribe",
		"unsubscribe",
		"get_presence",
	}, nil
}

func (c *EasyWsCapability) Handle(ctx *gin.Context) {
	token := ctx.Request.URL.Query().Get("token")
	if token == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claim, err := c.parseToken(token)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(ctx.Request, ctx.Writer)
	if err != nil {
		httpx.WriteErrString(ctx, "failed to upgrade websocket")
		return
	}

	_, err = c.room.AddConn(claim.UserId, conn, room.ConnId(claim.ResourceId))
	if err != nil {
		conn.Close()
		httpx.WriteErrString(ctx, "failed to add connection")
		return
	}
}

var ErrInvalidToken = errors.New("invalid token")

func (c *EasyWsCapability) parseToken(token string) (*signer.CapabilityClaim, error) {

	claim, err := c.builder.signer.ParseCapability(token)
	if err != nil {
		return nil, err
	}

	if claim.SpaceId != c.spaceId {
		return nil, ErrInvalidToken
	}

	if claim.InstallId != c.installId {
		return nil, ErrInvalidToken
	}

	if claim.CapabilityId != c.capabilityId {
		return nil, ErrInvalidToken
	}

	return claim, nil

}

func (c *EasyWsCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	// reload config

	return c, nil
}

func (c *EasyWsCapability) Close() error {
	// Cleanup can be done here if needed
	return nil
}
