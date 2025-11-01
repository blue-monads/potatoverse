package easyws

import (
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type EasyWsAddon struct {
	app     xtypes.App
	spaceId int64

	rooms map[string]*SockdRoom
	rLock sync.RWMutex
}

type SockdRoom struct {
	connections     map[int64]any
	connectionsLock sync.Mutex
}

type Conn interface {
	Id() int64
	Write([]byte) error
	Close() error
	Read() ([]byte, error)
}

func (e *EasyWsAddon) Name() string {
	return "easyws"
}

func (e *EasyWsAddon) Handle(ctx *gin.Context) error {

	return nil
}

func (e *EasyWsAddon) List() ([]string, error) {
	return nil, nil
}
func (e *EasyWsAddon) GetMeta(name string) (map[string]any, error) {
	return nil, nil
}
func (e *EasyWsAddon) Execute(method string, params xtypes.LazyData) (map[string]any, error) {
	return nil, nil
}
