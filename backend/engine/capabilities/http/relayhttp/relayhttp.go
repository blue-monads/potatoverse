package relayhttp

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

/*

relay http is a capability where one request uploads data [POST]
and another request downloads data [GET]

*/

const (
	bufferSize        = 32 * 1024 // 32KB
	channelBufferSize = 4
)

var (
	Name         = "relay-http"
	Icon         = "relay"
	OptionFields = []xcapability.CapabilityOptionField{}
)

type RelayHttpCapability struct {
	parent      *RelayHttpBuilder
	handle      xcapability.XCapabilityHandle
	allowSubKey bool
}

type RelayHttp struct {
	data   chan []byte
	closed bool
}

func (c *RelayHttpCapability) getOrCreateRelay(relayID string) *RelayHttp {
	c.parent.rLock.RLock()
	relay, exists := c.parent.httpRelays[relayID]
	c.parent.rLock.RUnlock()

	if exists {
		return relay
	}

	newrelay := &RelayHttp{
		data: make(chan []byte, channelBufferSize),
	}

	// Create new relay
	c.parent.rLock.Lock()
	defer c.parent.rLock.Unlock()

	// Double-check after acquiring write lock
	relay, exists = c.parent.httpRelays[relayID]
	if exists {
		return relay
	}

	c.parent.httpRelays[relayID] = newrelay

	return newrelay
}

func (c *RelayHttpCapability) removeRelay(relayID string) {
	c.parent.rLock.Lock()
	defer c.parent.rLock.Unlock()
	delete(c.parent.httpRelays, relayID)
}

func (c *RelayHttpCapability) Handle(ctx *gin.Context) {
	// Route based on HTTP method
	switch ctx.Request.Method {
	case "POST":
		c.HandleSender(ctx)
	case "GET":
		c.HandleReceiver(ctx)
	default:
		ctx.JSON(405, gin.H{"error": "Method not allowed"})
	}
}

func (c *RelayHttpCapability) getRelayID(ctx *gin.Context) (string, error) {

	token := ctx.GetHeader("Authorization")
	if token == "" {
		token = ctx.Query("token")
		if token == "" {
			return "", errors.New("token is required")
		}
	}

	claim, err := c.handle.ValidateCapToken(token)
	if err != nil {
		return "", err
	}

	// <capability_id>_<resource_id>:<sub_key>

	relayId := fmt.Sprintf("%d_%s", claim.CapabilityId, claim.ResourceId)
	if relayId == "" {
		return "", errors.New("relay id is required")
	}

	if c.allowSubKey {
		subKey := claim.ExtraMeta["sub_key"]
		if subKey == "" {
			return "", errors.New("sub key is required")
		}
		relayId = fmt.Sprintf("%s:%s", relayId, subKey)
	}

	return relayId, nil

}

func (c *RelayHttpCapability) HandleSender(ctx *gin.Context) {

	relayID, err := c.getRelayID(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get or create relay
	relay := c.getOrCreateRelay(relayID)

	// Stream data in chunks (blocking)
	reader := bufio.NewReaderSize(ctx.Request.Body, bufferSize)
	buf := make([]byte, bufferSize)

	defer func() {
		close(relay.data)

		if relay.closed {
			return
		}

		c.removeRelay(relayID)
		relay.closed = true

	}()

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			// Send chunk (copy to avoid reuse issues)
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			relay.data <- chunk
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			// Send error signal (nil chunk indicates error)
			relay.data <- nil
			close(relay.data)
			ctx.JSON(400, gin.H{"error": "Failed to read request body"})
			return
		}
	}

	ctx.JSON(200, gin.H{"status": "data relayed", "relay_id": relayID})
}

func (c *RelayHttpCapability) HandleReceiver(ctx *gin.Context) {

	relayID, err := c.getRelayID(ctx)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	relay := c.getOrCreateRelay(relayID)

	defer func() {
		if relay.closed {
			return
		}

		c.removeRelay(relayID)
		relay.closed = true
	}()

	// Set response headers for streaming
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.WriteHeader(200)

	rctx := ctx.Request.Context()

	for {
		select {
		case <-rctx.Done():
			return
		case chunk := <-relay.data:
			if chunk == nil {
				return
			}
			if _, err := ctx.Writer.Write(chunk); err != nil {
				return
			}
			ctx.Writer.Flush()
		}
	}

}

func (c *RelayHttpCapability) ListActions() ([]string, error) {
	return []string{}, nil
}

func (c *RelayHttpCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	return nil, errors.New("relay-http does not support actions")
}

func (c *RelayHttpCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return c, nil
}

func (c *RelayHttpCapability) Close() error {

	return nil
}
