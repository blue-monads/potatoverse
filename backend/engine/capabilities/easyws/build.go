package easyws

import (
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var Ok = struct {
	Success bool `json:"success"`
}{
	Success: true,
}

var (
	Name         = "easy-ws"
	Icon         = "socket"
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &EasyWsBuilder{
				app:    appTyped,
				rooms:  make(map[string]*room.Room),
				signer: appTyped.Signer(),
				engine: appTyped.Engine().(xtypes.Engine),
			}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type EasyWsBuilder struct {
	app    xtypes.App
	signer *signer.Signer

	engine xtypes.Engine

	rooms map[string]*room.Room
	rLock sync.Mutex
}

type CapabilityAccessHandle interface {
	ParseToken(token string) (*signer.CapabilityClaim, error)
	EmitActionEvent(opts *xtypes.ActionEventOptions) error
}

func (b *EasyWsBuilder) Build(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {

	roomName := fmt.Sprintf("cap-%d", model.ID)
	cmdChan := make(chan room.Message)
	onDisconnect := make(chan room.UserConnInfo)

	newRoom := room.NewRoom(room.Options{
		CmdChan:      cmdChan,
		OnDisconnect: onDisconnect,
	})

	b.rLock.Lock()
	defer b.rLock.Unlock()

	existingRoom := b.rooms[roomName]
	b.rooms[roomName] = newRoom

	if existingRoom != nil {
		existingRoom.Close()
	}

	ec := &EasyWsCapability{
		app:              b.app,
		spaceId:          model.SpaceID,
		installId:        model.InstallID,
		capabilityId:     model.ID,
		onCmdChan:        cmdChan,
		onDisconnectChan: onDisconnect,
		room:             newRoom,
	}

	go ec.evLoop()

	return ec, nil
}

func (b *EasyWsBuilder) Serve(ctx *gin.Context) {}

func (b *EasyWsBuilder) Name() string {
	return Name
}
