package eslayer

import (
	"context"
	"sync"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	qq "github.com/blue-monads/potatoverse/backend/utils/qq"
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

func NewESLayer(db datahub.Database, handlers map[string]evtype.Handler) *ESLayer {

	ctx, cancel := context.WithCancel(context.Background())

	return &ESLayer{
		datahandle:             db,
		handlers:               handlers,
		eventProcessChan:       make(chan int64, 20),
		eventTargetProcessChan: make(chan int64, 20),
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
	qq.Println("@NotifyNewEvent: called with event", eventId)
	select {
	case e.eventProcessChan <- eventId:
		qq.Println("@NotifyNewEvent: sent event", eventId)
	case <-e.ctx.Done():
		qq.Println("@NotifyNewEvent: context done")
	}
}
