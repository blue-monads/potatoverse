package easyws

import (
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type EasyWsCapability struct {
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

func (e *EasyWsCapability) Name() string {
	return "easyws"
}

func (e *EasyWsCapability) Handle(ctx *gin.Context) {
	// TODO: implement capability handling
}

func (e *EasyWsCapability) ListActions() ([]string, error) {
	return nil, nil
}

func (e *EasyWsCapability) GetActionMeta(name string) (map[string]any, error) {
	return nil, nil
}

func (e *EasyWsCapability) ExecuteAction(name string, params xtypes.LazyData) (map[string]any, error) {
	return nil, nil
}
