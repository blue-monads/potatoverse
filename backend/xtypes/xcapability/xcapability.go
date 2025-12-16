package xcapability

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
)

type Capability interface {
	Handle(ctx *gin.Context)
	ListActions() ([]string, error)
	Execute(name string, params lazydata.LazyData) (any, error)
	Reload(model *dbmodels.SpaceCapability) (Capability, error)
	Close() error
}

type CapabilityBuilderFactory struct {
	Name         string
	Icon         string
	OptionFields []CapabilityOptionField

	Builder func(app any) (CapabilityBuilder, error)
}

type CapabilityOptionField struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	// text, number, date, api_key, boolean, select, multi_select, textarea, object
	Type     string   `json:"type"`
	Default  string   `json:"default"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
}

type CapabilityBuilder interface {
	Name() string
	Build(model *dbmodels.SpaceCapability) (Capability, error)
	Serve(ctx *gin.Context)
}

type CapabilityHub interface {
	List(spaceId int64) ([]string, error)
	Execute(installId, spaceId int64, gname, method string, params lazydata.LazyData) (any, error)
	Methods(installId, spaceId int64, gname string) ([]string, error)
}
