package ccurd

import (
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type PingCapability struct {
	app     xtypes.App
	spaceId int64
}

func (p *PingCapability) Reload(opts xtypes.LazyData) (xtypes.Capability, error) {
	return p, nil
}

func (p *PingCapability) Close() error {
	return nil
}

func (p *PingCapability) Handle(ctx *gin.Context) {}

func (p *PingCapability) ListActions() ([]string, error) {
	return []string{}, nil
}

func (p *PingCapability) Execute(name string, params xtypes.LazyData) (map[string]any, error) {

	return nil, nil
}
