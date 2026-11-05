package xcapability

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
)

type Capability interface {
	Handle(ctx *gin.Context)
	ListActions() ([]string, error)
	Execute(ctx CapExecContext) (any, error)
	Reload(model *dbmodels.SpaceCapability) (Capability, error)
	Close() error
}

type CapabilityBuilderFactory struct {
	Name             string
	Icon             string
	FreeFieldOptions bool
	OptionFields     []CapabilityOptionField

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
	Build(handle XCapabilityHandle) (Capability, error)
	Serve(ctx *gin.Context)
	GetDebugData() map[string]any
}

type CapExecContext struct {
	State       *executors.ExecState
	HttpRequest *gin.Context

	CapabilityName string
	Method         string
	Params         lazydata.LazyData
}

type CapabilityHub interface {
	List(spaceId int64) ([]string, error)
	Methods(installId, spaceId int64, gname string) ([]string, error)
	Execute(ctx CapExecContext) (any, error)
}
