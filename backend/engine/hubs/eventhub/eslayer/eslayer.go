package eslayer

import (
	"context"
	"sync"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type ESLayer struct {
	datahandle evtype.DataHandle

	handlers map[string]evtype.Handler

	eventProcessChan       chan int64
	eventTargetProcessChan chan int64

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewESLayer(app xtypes.App) *ESLayer {

	db := app.Database()

	ctx, cancel := context.WithCancel(context.Background())

	return &ESLayer{
		datahandle:             db,
		eventProcessChan:       make(chan int64, 13),
		eventTargetProcessChan: make(chan int64, 27),
		ctx:                    ctx,
		cancel:                 cancel,
		wg:                     sync.WaitGroup{},
	}
}

func (e *ESLayer) Start() error {
	e.eventLoop()
	return nil
}

func (e *ESLayer) Stop() {
	e.cancel()
	e.wg.Wait()
}

func (e *ESLayer) NotifyNewEvent(eventId int64) {

}
