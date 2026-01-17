package corehub

import (
	"encoding/json"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

func (c *CoreHub) UserSendMessage(msg *dbmodels.UserMessage) (int64, error) {
	notifier := c.sockd.GetNotifier()

	now := time.Now()
	msg.CreatedAt = &now

	id, err := c.db.GetUserOps().AddUserMessage(msg)
	if err != nil {
		return 0, err
	}
	msg.ID = id

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}

	err = notifier.SendUser(msg.ToUser, jsonMsg)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (c *CoreHub) HandleUserWS(userId int64, ctx *gin.Context) {
	dbUser, err := c.db.GetUserOps().GetUser(userId)
	if err != nil {
		httpx.WriteErrString(ctx, "failed to get user")
		return
	}

	if dbUser.IsDeleted {
		httpx.WriteErrString(ctx, "user is deleted")
		return
	}

	if !dbUser.IsVerified {
		httpx.WriteErrString(ctx, "user is not verified")
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(ctx.Request, ctx.Writer)
	if err != nil {
		httpx.WriteErrString(ctx, "failed to upgrade websocket")
		return
	}

	ugroup := dbUser.Ugroup

	connId, err := c.sockd.GetNotifier().AddUserConnection(userId, ugroup, conn)
	if err != nil {
		httpx.WriteErrString(ctx, "failed to add user connection")
		return
	}

	qq.Println("@HandleUserWS/user connected", userId, connId)

}
