package ping

import (
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "Ping"
	Icon         = "ping"
	OptionFields = []xtypes.CapabilityOptionField{
		{
			Name:        "Add Radom number to the result",
			Key:         "add_random_number",
			Description: "Add a random number to the result",
			Type:        "boolean",
			Default:     "false",
		},

		// example fields for all field types
		{
			Name:        "Text",
			Key:         "text",
			Description: "Text field",
			Type:        "text",
			Default:     "",
		},
		{
			Name:        "Number",
			Key:         "number",
			Description: "Number field",
			Type:        "number",
			Default:     "0",
		},

		{
			Name:        "Date",
			Key:         "date",
			Description: "Date field",
			Type:        "date",
			Default:     "",
		},

		{
			Name:        "API Key",
			Key:         "api_key",
			Description: "API Key field",
			Type:        "api_key",
			Default:     "",
		},

		{
			Name:        "Select",
			Key:         "select",
			Description: "Select field",
			Type:        "select",
			Default:     "",
			Options:     []string{"option1", "option2", "option3"},
		},

		{
			Name:        "Multi Select",
			Key:         "multi_select",
			Description: "Multi Select field",
			Type:        "multi_select",
			Default:     "",
			Options:     []string{"option1", "option2", "option3"},
		},

		{
			Name:        "Textarea",
			Key:         "textarea",
			Description: "Textarea field",
			Type:        "textarea",
			Default:     "",
		},

		{
			Name:        "Object",
			Key:         "object",
			Description: "Object field",
			Type:        "object",
			Default:     "{}",
		},
	}
)

func init() {
	registry.RegisterCapability("ping", xtypes.CapabilityBuilderFactory{
		Builder: func(app xtypes.App) (xtypes.CapabilityBuilder, error) {
			return &PingBuilder{app: app}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type PingBuilder struct {
	app xtypes.App
}

func (b *PingBuilder) Build(model *dbmodels.SpaceCapability) (xtypes.Capability, error) {
	return &PingCapability{
		app:     b.app,
		spaceId: model.SpaceID,
	}, nil
}

func (b *PingBuilder) Serve(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "pong",
		"capability": "ping",
	})
}

func (p *PingBuilder) Name() string {
	return "ping"
}

type PingCapability struct {
	app     xtypes.App
	spaceId int64
}

func (p *PingCapability) Reload(model *dbmodels.SpaceCapability) (xtypes.Capability, error) {
	return p, nil
}

func (p *PingCapability) Close() error {
	return nil
}

func (p *PingCapability) Handle(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "pong",
		"capability": "ping",
		"space_id":   p.spaceId,
	})
}

func (p *PingCapability) ListActions() ([]string, error) {
	return []string{"ping"}, nil
}

func (p *PingCapability) Execute(name string, params xtypes.LazyData) (map[string]any, error) {
	if name == "ping" {
		return map[string]any{
			"result":   "pong",
			"space_id": p.spaceId,
		}, nil
	}
	return nil, nil
}
