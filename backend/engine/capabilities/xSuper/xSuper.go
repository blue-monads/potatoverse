package xsuper

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xSuper"
	Icon         = "<i class='fa-solid fa-shield-halved'></i>"
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &SuperBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type SuperBuilder struct {
	app xtypes.App
}

func (b *SuperBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()
	return &SuperCapability{
		app:        b.app,
		handle:     handle,
		spaceId:    model.SpaceID,
		installId:  model.InstallID,
		capability: model,
	}, nil
}

func (b *SuperBuilder) Serve(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "super capability",
		"capability": Name,
	})
}

func (b *SuperBuilder) Name() string {
	return Name
}

func (b *SuperBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}

type SuperCapability struct {
	app        xtypes.App
	handle     xcapability.XCapabilityHandle
	spaceId    int64
	installId  int64
	capability *dbmodels.SpaceCapability
}

func (c *SuperCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	c.capability = model
	return c, nil
}

func (c *SuperCapability) Close() error {
	return nil
}

func (c *SuperCapability) Handle(ctx *gin.Context) {
	action := ctx.Param("action")
	switch action {
	case "get_admin_token":
		res, err := c.getAdminToken(nil)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, res)
	case "get_instance_info":
		res, err := c.getInstanceInfo(nil)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, res)
	default:
		ctx.JSON(200, gin.H{
			"message":    "super capability",
			"capability": Name,
			"space_id":   c.spaceId,
		})
	}
}

func (c *SuperCapability) ListActions() ([]string, error) {
	return []string{
		"get_admin_token",
		"get_instance_info",
	}, nil
}

func (c *SuperCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "get_admin_token":
		return c.getAdminToken(params)
	case "get_instance_info":
		return c.getInstanceInfo(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *SuperCapability) getAdminToken(params lazydata.LazyData) (any, error) {
	ctrl, ok := c.app.Controller().(*actions.Controller)
	if !ok || ctrl == nil {
		return nil, errors.New("controller not available")
	}

	user, token, err := ctrl.GetAdminToken()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"user":  user,
		"token": token,
	}, nil
}

func (c *SuperCapability) getInstanceInfo(params lazydata.LazyData) (any, error) {
	opts, ok := c.app.Config().(*xtypes.AppOptions)
	if !ok || opts == nil {
		return nil, errors.New("app config not available")
	}

	hosts := make([]string, 0, len(opts.Hosts))
	for _, h := range opts.Hosts {
		hosts = append(hosts, h.Name)
	}

	var host string
	if len(hosts) > 0 {
		host = hosts[0]
	}

	return map[string]any{
		"host":  host,
		"hosts": hosts,
		"port":  opts.Port,
	}, nil
}
