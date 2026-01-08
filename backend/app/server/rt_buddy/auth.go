package rtbuddy

import (
	"fmt"
	"time"

	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
)

func verifyNostrAuthCtx(ctx *gin.Context, expiry time.Duration) (*nostr.Event, error) {
	authHeader := ctx.GetHeader("X-Buddy-Auth")
	if authHeader == "" {
		return nil, fmt.Errorf("Unauthorized")
	}

	event, err := xutils.VerifyNostrAuth(authHeader)
	if err != nil {
		return nil, err
	}

	// check expiry
	if event.CreatedAt < nostr.Timestamp(time.Now().Add(-expiry).Unix()) {
		return nil, fmt.Errorf("Expired")
	}

	return event, nil

}
