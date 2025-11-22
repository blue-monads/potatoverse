package ccurd

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type PingCapability struct {
	spaceId int64
	db      datahub.DBLowOps
	signer  *signer.Signer
	methods map[string]*Methods
}

func (p *PingCapability) Reload(opts xtypes.LazyData) (xtypes.Capability, error) {
	return p, nil
}

func (p *PingCapability) Close() error {
	return nil
}

// POST /ccurd/methods
func (p *PingCapability) Handle(ctx *gin.Context) {
	token := ctx.Request.Header.Get("x-cap-token")
	if token == "" {
		httpx.WriteErrString(ctx, "Empty token")
		return
	}

	claim, err := p.signer.ParseCapability(token)
	if err != nil {
		httpx.WriteErrString(ctx, "token error")
		return
	}

	if claim.SpaceId != p.spaceId {
		httpx.WriteErrString(ctx, "token `error")
	}

	// if ctx.Request.Method == "POST" {
	// 	sdb := p.app.Database().GetLowCapabilityDBOps(fmt.Sprint(p.spaceId))

	// }

}

func (p *PingCapability) ListActions() ([]string, error) {
	return []string{}, nil
}

func (p *PingCapability) Execute(name string, params xtypes.LazyData) (map[string]any, error) {

	return nil, nil
}
