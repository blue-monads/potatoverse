package sighub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability/easyaction"
)

type SigHub struct {
	app    xtypes.App
	sigOps datahub.SignalOps

	activeSignals     map[string][]dbmodels.SignalLite
	activeSignalsLock sync.RWMutex

	refreshFullIndex  chan struct{}
	signalProcessChan chan int64

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewSigHub(app xtypes.App) *SigHub {
	db := app.Database()
	sigOps := db.GetSignalOps()

	ctx, cancel := context.WithCancel(context.Background())

	return &SigHub{
		app:               app,
		sigOps:            sigOps,
		activeSignals:     make(map[string][]dbmodels.SignalLite),
		activeSignalsLock: sync.RWMutex{},
		refreshFullIndex:  make(chan struct{}, 1),
		signalProcessChan: make(chan int64, 50),
		ctx:               ctx,
		cancel:            cancel,
		wg:                sync.WaitGroup{},
	}
}

func (s *SigHub) Start() error {
	err := s.buildActiveSignalsIndex()
	if err != nil {
		qq.Println("@SigHub/Start/buildActiveSignalsIndex/error", err)
		return err
	}

	go s.rootWatcher()
	go s.watchReload()

	// Start worker pool for signal events
	for range 5 {
		go s.workerLoop()
	}

	return nil
}

func (s *SigHub) Stop() {
	if s == nil {
		return
	}
	s.cancel()
	s.wg.Wait()
}

func (s *SigHub) RefreshFullIndex() {
	select {
	case s.refreshFullIndex <- struct{}{}:
	default:
	}
}

func (s *SigHub) Publish(opts *xtypes.SignalOptions) error {
	if opts == nil {
		return nil
	}

	qq.Println("@SigHub/Publish", opts.SignalKey, opts.EmitterInstallId, opts.EmitterSpaceId)

	matchingSignals := s.getMatchingSignals(opts.EmitterInstallId, opts.EmitterSpaceId, opts.SignalKey)
	if len(matchingSignals) == 0 {
		qq.Println("@SigHub/Publish: no matching active signals for", opts.SignalKey)
		return nil
	}

	for _, sig := range matchingSignals {
		eventId, err := s.sigOps.AddSignalEvent(sig.ID, opts.Payload, opts.Metadata)
		if err != nil {
			qq.Println("@SigHub/Publish/AddSignalEvent/error", err)
			continue
		}

		s.notifyNewEvent(eventId)
	}

	return nil
}

func (s *SigHub) notifyNewEvent(eventId int64) {
	select {
	case s.signalProcessChan <- eventId:
	case <-s.ctx.Done():
	}
}

func (s *SigHub) buildActiveSignalsIndex() error {
	signals, err := s.sigOps.QueryAllActiveSignals()
	if err != nil {
		return err
	}

	nextIndex := make(map[string][]dbmodels.SignalLite)
	for _, sig := range signals {
		key := fmt.Sprintf("%d||%s", sig.EmitterInstallID, sig.SignalKey)
		nextIndex[key] = append(nextIndex[key], sig)
	}

	s.activeSignalsLock.Lock()
	defer s.activeSignalsLock.Unlock()

	s.activeSignals = nextIndex
	return nil
}

func (s *SigHub) getMatchingSignals(emitterInstallId, emitterSpaceId int64, signalKey string) []dbmodels.SignalLite {
	key := fmt.Sprintf("%d||%s", emitterInstallId, signalKey)

	s.activeSignalsLock.RLock()
	defer s.activeSignalsLock.RUnlock()

	signals := s.activeSignals[key]
	if len(signals) == 0 {
		return nil
	}

	now := time.Now().Unix()
	var matched []dbmodels.SignalLite
	for _, sig := range signals {
		if sig.Disabled {
			continue
		}
		if sig.ExpiresOn > 0 && sig.ExpiresOn < now {
			continue
		}
		if sig.EmitterSpaceID != 0 && emitterSpaceId != 0 && sig.EmitterSpaceID != emitterSpaceId {
			continue
		}
		matched = append(matched, sig)
	}
	return matched
}

func (s *SigHub) watchReload() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.refreshFullIndex:
			time.Sleep(100 * time.Millisecond)
			err := s.buildActiveSignalsIndex()
			if err != nil {
				qq.Println("@SigHub/watchReload/error", err)
			}
		}
	}
}

func (s *SigHub) rootWatcher() {
	s.wg.Add(1)
	defer s.wg.Done()

	checkForNew := func() {
		eventIds, err := s.sigOps.QueryNewSignalEvents()
		if err != nil {
			qq.Println("@SigHub/rootWatcher/QueryNewSignalEvents/error", err)
			return
		}
		for _, id := range eventIds {
			s.notifyNewEvent(id)
		}
	}

	checkForDelayed := func() {
		eventIds, err := s.sigOps.QueryDelayExpiredSignalEvents()
		if err != nil {
			qq.Println("@SigHub/rootWatcher/QueryDelayExpiredSignalEvents/error", err)
			return
		}
		for _, id := range eventIds {
			s.notifyNewEvent(id)
		}
	}

	timer := time.NewTicker(15 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-timer.C:
			checkForNew()
			checkForDelayed()
		}
	}
}

func (s *SigHub) workerLoop() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case eventId, ok := <-s.signalProcessChan:
			if !ok {
				return
			}
			if eventId == 0 {
				continue
			}
			s.processSignalEvent(eventId)
		}
	}
}

func (s *SigHub) processSignalEvent(eventId int64) {
	evt, err := s.sigOps.GetSignalEvent(eventId)
	if err != nil {
		qq.Println("@SigHub/processSignalEvent/GetSignalEvent/error", err)
		return
	}

	if evt.Status == "processed" || evt.Status == "expired" {
		return
	}

	sig, err := s.sigOps.GetSignal(evt.SignalID)
	if err != nil {
		qq.Println("@SigHub/processSignalEvent/GetSignal/error", err)
		s.sigOps.TransitionSignalEventFail(eventId, fmt.Sprintf("signal not found: %v", err))
		return
	}

	if sig.Disabled {
		qq.Println("@SigHub/processSignalEvent: signal disabled", sig.ID)
		return
	}

	// Check if signal has expired
	now := time.Now().Unix()
	if sig.ExpiresOn > 0 && sig.ExpiresOn < now {
		qq.Println("@SigHub/processSignalEvent: signal expired", sig.ID)
		s.sigOps.TransitionSignalEventExpired(eventId)
		return
	}

	// Transition to processing
	_, err = s.sigOps.TransitionSignalEventStart(eventId)
	if err != nil {
		qq.Println("@SigHub/processSignalEvent/TransitionSignalEventStart/error", err)
		return
	}

	// Dispatch to receiver space via Engine
	engine, ok := s.app.Engine().(xtypes.Engine)
	if !ok {
		err := fmt.Errorf("app engine does not implement xtypes.Engine")
		s.sigOps.TransitionSignalEventFail(eventId, err.Error())
		return
	}

	actionErr := engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    sig.ReceiverSpaceID,
		EventType:  "signal",
		ActionName: sig.ReceiverHandler,
		Params: map[string]string{
			"signal_key":          sig.SignalKey,
			"signal_id":           fmt.Sprintf("%d", sig.ID),
			"signal_event_id":     fmt.Sprintf("%d", evt.ID),
			"emitter_install_id":  fmt.Sprintf("%d", sig.EmitterInstallID),
			"emitter_space_id":    fmt.Sprintf("%d", sig.EmitterSpaceID),
			"receiver_install_id": fmt.Sprintf("%d", sig.ReceiverInstallID),
			"receiver_space_id":   fmt.Sprintf("%d", sig.ReceiverSpaceID),
		},
		Request: &easyaction.Context{
			Payload: evt.Payload,
		},
	})

	if actionErr != nil {
		qq.Println("@SigHub/processSignalEvent/actionErr", actionErr)
		// Check retry policy
		if sig.MaxRetries > 0 && evt.RetryCount < sig.MaxRetries {
			delayUntil := time.Now().Unix() + sig.RetryDelay
			_ = s.sigOps.TransitionSignalEventDelay(eventId, delayUntil, evt.RetryCount+1, actionErr.Error())
			qq.Println("@SigHub/processSignalEvent: scheduled retry", evt.RetryCount+1, "at", delayUntil)
			return
		}

		_ = s.sigOps.TransitionSignalEventFail(eventId, actionErr.Error())
		return
	}

	_ = s.sigOps.TransitionSignalEventComplete(eventId)
	qq.Println("@SigHub/processSignalEvent: completed", eventId)
}
