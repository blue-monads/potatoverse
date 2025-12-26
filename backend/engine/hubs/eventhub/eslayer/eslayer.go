package eslayer

import (
	"context"
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type ESLayer struct {
	app  xtypes.App
	sink datahub.MQSynk
	db   datahub.Database

	eventProcessChan       chan int64
	eventTargetProcessChan chan int64

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewESLayer(app xtypes.App) *ESLayer {

	db := app.Database()
	sink := db.GetMQSynk()

	ctx, cancel := context.WithCancel(context.Background())

	return &ESLayer{
		app:                    app,
		sink:                   sink,
		db:                     db,
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
