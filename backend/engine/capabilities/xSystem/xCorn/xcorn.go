package xcorn

import (
	"sync"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type CornCapability struct {
	builder *CornBuilder
	handle  xcapability.XCapabilityHandle

	jobs      map[string]*CornJob
	reload    chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

func (c *CornCapability) Handle(ctx *gin.Context) {}

func (c *CornCapability) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
	})
	return nil
}

func (c *CornCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	newCap, err := c.builder.Build(c.handle)
	if err != nil {
		return nil, err
	}

	c.done <- struct{}{}

	return newCap, nil
}
