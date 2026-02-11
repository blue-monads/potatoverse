package rtbuddy

import (
	"net/http"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

func (a *BuddyRouteServer) handleBuddyPing(ctx *gin.Context) {

	pubkey, err := verifyNostrAuthCtx(ctx, BuddyAuthExpiry)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	qq.Println("@buddy_pubkey", pubkey)

	serverPubkey := a.buddyhub.GetPubkey()

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Buddy pinged",
		"server_pubkey": serverPubkey,
	})

}
