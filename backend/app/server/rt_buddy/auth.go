package rtbuddy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
)

func verifyNostrAuthCtx(ctx *gin.Context, expiry time.Duration) (*nostr.Event, error) {
	authHeader := ctx.GetHeader("X-Buddy-Auth")
	if authHeader == "" {
		return nil, fmt.Errorf("Unauthorized")
	}

	event, err := verifyNostrAuth(authHeader)
	if err != nil {
		return nil, err
	}

	// check expiry
	if event.CreatedAt < nostr.Timestamp(time.Now().Add(-expiry).Unix()) {
		return nil, fmt.Errorf("Expired")
	}

	return event, nil

}

func verifyNostrAuth(authHeader string) (*nostr.Event, error) {

	eventJson, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	var event nostr.Event
	err = json.Unmarshal(eventJson, &event)
	if err != nil {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	ok, err := event.CheckSignature()
	if !ok || err != nil {
		return nil, fmt.Errorf("invalid signature")
	}

	if event.Kind != nostr.KindHTTPAuth {
		return nil, fmt.Errorf("wrong event kind")
	}

	return &event, nil
}
