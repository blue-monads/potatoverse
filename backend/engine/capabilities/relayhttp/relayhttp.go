package relayhttp

import (
	"bufio"
	"io"
	"sync"

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

type RelayHttpCapability struct {
	httpRelays map[string]*RelayHttp
	rLock      sync.RWMutex
}

type RelayHttp struct {
	data chan []byte
}

func (c *RelayHttpCapability) getOrCreateRelay(relayID string) *RelayHttp {
	c.rLock.RLock()
	relay, exists := c.httpRelays[relayID]
	c.rLock.RUnlock()

	if exists {
		return relay
	}

	newrelay := &RelayHttp{
		data: make(chan []byte, channelBufferSize),
	}

	// Create new relay
	c.rLock.Lock()
	defer c.rLock.Unlock()

	// Double-check after acquiring write lock
	relay, exists = c.httpRelays[relayID]
	if exists {
		return relay
	}

	c.httpRelays[relayID] = newrelay

	return newrelay
}

func (c *RelayHttpCapability) removeRelay(relayID string) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	delete(c.httpRelays, relayID)
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

func (c *RelayHttpCapability) HandleSender(ctx *gin.Context) {
	// Extract relay ID from query parameter or path
	relayID := ctx.Query("relay_id")
	if relayID == "" {
		relayID = ctx.Param("relay_id")
	}
	if relayID == "" {
		ctx.JSON(400, gin.H{"error": "relay_id is required"})
		return
	}

	// Get or create relay
	relay := c.getOrCreateRelay(relayID)

	// Stream data in chunks (blocking)
	reader := bufio.NewReaderSize(ctx.Request.Body, bufferSize)
	buf := make([]byte, bufferSize)

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

	// Close channel to signal end of stream
	close(relay.data)

	ctx.JSON(200, gin.H{"status": "data relayed", "relay_id": relayID})
}

func (c *RelayHttpCapability) HandleReceiver(ctx *gin.Context) {
	// Extract relay ID from query parameter or path
	relayID := ctx.Query("relay_id")
	if relayID == "" {
		relayID = ctx.Param("relay_id")
	}
	if relayID == "" {
		ctx.JSON(400, gin.H{"error": "relay_id is required"})
		return
	}

	relay := c.getOrCreateRelay(relayID)

	defer c.removeRelay(relayID)

	// Set response headers for streaming
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
	ctx.Writer.WriteHeader(200)

	for chunk := range relay.data {
		if chunk == nil {
			ctx.JSON(500, gin.H{"error": "error during data transfer"})
			return
		}

		if _, err := ctx.Writer.Write(chunk); err != nil {
			return
		}

		ctx.Writer.Flush()
	}

}
