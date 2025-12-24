package easyws

import (
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/capabilities/websocket/easyws/room"
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
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
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name: "on_connect_action",
			Key:  "on_connect_action",
			Type: "boolean",
		},
		{
			Name: "on_disconnect_action",
			Key:  "on_disconnect_action",
			Type: "boolean",
		},
		{
			Name: "on_command_action",
			Key:  "on_command_action",
			Type: "boolean",
		},
	}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			builder := &EasyWsBuilder{
				app:              appTyped,
				rooms:            make(map[string]*room.Room),
				signer:           appTyped.Signer(),
				engine:           appTyped.Engine().(xtypes.Engine),
				onCmdChan:        make(chan CMDMessage),
				onDisconnectChan: make(chan CMDDisconnectMessage),
				rLock:            sync.Mutex{},
			}
			go builder.evLoop()

			return builder, nil
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

	onCmdChan        chan CMDMessage
	onDisconnectChan chan CMDDisconnectMessage
}

type CapabilityAccessHandle interface {
	ParseToken(token string) (*signer.CapabilityClaim, error)
	EmitActionEvent(opts *xtypes.ActionEventOptions) error
}

func (b *EasyWsBuilder) Build(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {

	opt := lazydata.LazyDataBytes(kosher.Byte(model.Options))

	onConnectAction := opt.GetFieldAsBool("on_connect_action")
	onDisconnectAction := opt.GetFieldAsBool("on_disconnect_action")
	onCommandAction := opt.GetFieldAsBool("on_command_action")

	ec := &EasyWsCapability{
		builder:         b,
		spaceId:         model.SpaceID,
		installId:       model.InstallID,
		capabilityId:    model.ID,
		room:            nil,
		onConnectAction: onConnectAction,
	}

	roomName := fmt.Sprintf("cap-%d", model.ID)

	var onCommand func(msg room.CommandMessage) error
	var onDisconnect func(msg room.DisconnectMessage) error

	if onCommandAction {
		onCommand = func(msg room.CommandMessage) error {

			b.onCmdChan <- CMDMessage{
				c:   ec,
				cmd: msg,
			}
			return nil
		}
	}

	if onDisconnectAction {
		onDisconnect = func(msg room.DisconnectMessage) error {

			qq.Println("@Build/onDisconnect", msg.ConnId, msg.UserId)

			b.onDisconnectChan <- CMDDisconnectMessage{
				c:   ec,
				msg: msg,
			}
			return nil
		}
	}

	newRoom := room.NewRoom(room.Options{

		OnCommand: onCommand,

		OnDisconnect: onDisconnect,
	})

	b.rLock.Lock()
	defer b.rLock.Unlock()

	existingRoom := b.rooms[roomName]
	b.rooms[roomName] = newRoom

	if existingRoom != nil {
		existingRoom.Close()
	}

	ec.room = newRoom

	go newRoom.Run()

	return ec, nil
}

func (b *EasyWsBuilder) Serve(ctx *gin.Context) {}

func (b *EasyWsBuilder) Name() string {
	return Name
}

func (b *EasyWsBuilder) GetDebugData() map[string]any {
	b.rLock.Lock()
	defer b.rLock.Unlock()

	rooms := make(map[string]any)
	for name, room := range b.rooms {
		rooms[name] = room.GetDebugData()
	}

	return map[string]any{
		"rooms": rooms,
	}
}
