package easyws

import (
	"errors"
	"net/http"
	"strings"

	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

type EasyWsCapability struct {
	builder      *EasyWsBuilder
	spaceId      int64
	installId    int64
	capabilityId int64
	room         *room.Room

	onConnectAction bool
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

	if strings.HasSuffix(ctx.Request.URL.Path, "/test") {
		c.HandleTest(ctx)
		return
	}

	token := ctx.Request.URL.Query().Get("token")
	if token == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claim, err := c.parseToken(token)
	if err != nil {
		qq.Println("failed to parse token: ", err)

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

	if c.onConnectAction {
		err = c.afterConnect(claim.ResourceId, claim.UserId)
		if err != nil {
			httpx.WriteErrString(ctx, "failed to execute after_connect action")
			return
		}
	}

}

func (c *EasyWsCapability) HandleTest(ctx *gin.Context) {
	claim := &signer.CapabilityClaim{
		SpaceId:      c.spaceId,
		InstallId:    c.installId,
		CapabilityId: c.capabilityId,
		UserId:       1,
		ResourceId:   "test",
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

	if c.onConnectAction {
		err = c.afterConnect(claim.ResourceId, claim.UserId)
		if err != nil {
			httpx.WriteErrString(ctx, "failed to execute after_connect action")
			return
		}
	}

}

var ErrInvalidToken = errors.New("invalid token")

func (c *EasyWsCapability) parseToken(token string) (*signer.CapabilityClaim, error) {

	claim, err := c.builder.signer.ParseCapability(token)
	if err != nil {
		qq.Println("failed to parse token: ", err)
		return nil, err
	}

	if claim.SpaceId != c.spaceId {
		qq.Println("invalid space id: ", claim.SpaceId, "expected: ", c.spaceId)
		return nil, ErrInvalidToken
	}

	if claim.InstallId != c.installId {
		qq.Println("invalid install id: ", claim.InstallId, "expected: ", c.installId)
		return nil, ErrInvalidToken
	}

	if claim.CapabilityId != c.capabilityId {
		qq.Println("invalid capability id: ", claim.CapabilityId, "expected: ", c.capabilityId)
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
