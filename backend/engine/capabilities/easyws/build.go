package easyws

import (
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
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
	OptionFields = []xtypes.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(Name, xtypes.CapabilityBuilderFactory{
		Builder: func(app xtypes.App) (xtypes.CapabilityBuilder, error) {
			return &EasyWsBuilder{
				app:    app,
				rooms:  make(map[string]*room.Room),
				signer: app.Signer(),
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

	rooms map[string]*room.Room
	rLock sync.Mutex
}

func (b *EasyWsBuilder) Build(model *dbmodels.SpaceCapability) (xtypes.Capability, error) {

	roomName := fmt.Sprintf("cap-%d", model.ID)
	cmdChan := make(chan room.Message)
	newRoom := room.NewRoom(cmdChan)

	b.rLock.Lock()
	defer b.rLock.Unlock()

	existingRoom := b.rooms[roomName]
	b.rooms[roomName] = newRoom

	if existingRoom != nil {
		existingRoom.Close()
	}

	ec := &EasyWsCapability{
		app:          b.app,
		spaceId:      model.SpaceID,
		installId:    model.InstallID,
		capabilityId: model.ID,
		cmdChan:      cmdChan,
		room:         newRoom,
	}

	go ec.handleCommand()

	return ec, nil
}

func (b *EasyWsBuilder) Serve(ctx *gin.Context) {}

func (b *EasyWsBuilder) Name() string {
	return Name
}
