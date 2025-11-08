package ping

import (
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

func init() {
	registry.RegisterCapability("ping", xtypes.CapabilityBuilderFactory{
		Builder: func(app xtypes.App) (xtypes.CapabilityBuilder, error) {
			return &PingBuilder{app: app}, nil
		},
		Name: "Ping",
		Icon: "ping",
		OptionFields: []xtypes.CapabilityOptionField{
			{
				Name:        "Add Radom number to the result",
				Key:         "add_random_number",
				Description: "Add a random number to the result",
				Type:        "boolean",
				Default:     "false",
			},
		},
	})
}

type PingBuilder struct {
	app xtypes.App
}

func (b *PingBuilder) Build(spaceId int64) (xtypes.Capability, error) {
	return &PingCapability{
		app:     b.app,
		spaceId: spaceId,
	}, nil
}

func (b *PingBuilder) Serve(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message":    "pong",
		"capability": "ping",
	})
}

type PingCapability struct {
	app     xtypes.App
	spaceId int64
}

func (p *PingCapability) Name() string {
	return "ping"
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

func (p *PingCapability) GetActionMeta(name string) (map[string]any, error) {
	if name == "ping" {
		return map[string]any{
			"name":        "ping",
			"description": "Returns pong",
			"returns":     "string",
		}, nil
	}
	return nil, nil
}

func (p *PingCapability) ExecuteAction(name string, params xtypes.LazyData) (map[string]any, error) {
	if name == "ping" {
		return map[string]any{
			"result":   "pong",
			"space_id": p.spaceId,
		}, nil
	}
	return nil, nil
}
