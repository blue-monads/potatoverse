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
	targetProcessChan chan int64

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
		targetProcessChan: make(chan int64, 50),
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

	// Start worker pool for signal targets
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

	// Store payload once in SignalEvents
	eventId, err := s.sigOps.AddSignalEvent(opts.SignalKey, opts.Payload, opts.Metadata)
	if err != nil {
		qq.Println("@SigHub/Publish/AddSignalEvent/error", err)
		return err
	}

	// Create target entry for each matching signal
	for _, sig := range matchingSignals {
		targetId, err := s.sigOps.AddSignalTarget(eventId, sig.ID, opts.Metadata)
		if err != nil {
			qq.Println("@SigHub/Publish/AddSignalTarget/error", err)
			continue
		}

		s.notifyNewTarget(targetId)
	}

	return nil
}

func (s *SigHub) notifyNewTarget(targetId int64) {
	select {
	case s.targetProcessChan <- targetId:
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
		targetIds, err := s.sigOps.QueryNewSignalTargets()
		if err != nil {
			qq.Println("@SigHub/rootWatcher/QueryNewSignalTargets/error", err)
			return
		}
		for _, id := range targetIds {
			s.notifyNewTarget(id)
		}
	}

	checkForDelayed := func() {
		targetIds, err := s.sigOps.QueryDelayExpiredSignalTargets()
		if err != nil {
			qq.Println("@SigHub/rootWatcher/QueryDelayExpiredSignalTargets/error", err)
			return
		}
		for _, id := range targetIds {
			s.notifyNewTarget(id)
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
		case targetId, ok := <-s.targetProcessChan:
			if !ok {
				return
			}
			if targetId == 0 {
				continue
			}
			s.processSignalTarget(targetId)
		}
	}
}

func (s *SigHub) processSignalTarget(targetId int64) {
	tgt, err := s.sigOps.GetSignalTarget(targetId)
	if err != nil {
		qq.Println("@SigHub/processSignalTarget/GetSignalTarget/error", err)
		return
	}

	if tgt.Status == "processed" || tgt.Status == "expired" || tgt.Status == "blocked" {
		return
	}

	sig, err := s.sigOps.GetSignal(tgt.SignalID)
	if err != nil {
		qq.Println("@SigHub/processSignalTarget/GetSignal/error", err)
		_ = s.sigOps.TransitionSignalTargetFail(targetId, fmt.Sprintf("signal not found: %v", err))
		_, _ = s.sigOps.CheckAndCleanupSignalEvent(tgt.SignalEventID)
		return
	}

	if sig.Disabled {
		qq.Println("@SigHub/processSignalTarget: signal disabled", sig.ID)
		return
	}

	// Check if signal has expired
	now := time.Now().Unix()
	if sig.ExpiresOn > 0 && sig.ExpiresOn < now {
		qq.Println("@SigHub/processSignalTarget: signal expired", sig.ID)
		_ = s.sigOps.TransitionSignalTargetExpired(targetId)
		_, _ = s.sigOps.CheckAndCleanupSignalEvent(tgt.SignalEventID)
		return
	}

	evt, err := s.sigOps.GetSignalEvent(tgt.SignalEventID)
	if err != nil {
		qq.Println("@SigHub/processSignalTarget/GetSignalEvent/error", err)
		_ = s.sigOps.TransitionSignalTargetFail(targetId, fmt.Sprintf("event not found: %v", err))
		return
	}

	// Transition to scheduled
	_, err = s.sigOps.TransitionSignalTargetStart(targetId)
	if err != nil {
		qq.Println("@SigHub/processSignalTarget/TransitionSignalTargetStart/error", err)
		return
	}

	// Dispatch to receiver space via Engine
	engine, ok := s.app.Engine().(xtypes.Engine)
	if !ok {
		err := fmt.Errorf("app engine does not implement xtypes.Engine")
		_ = s.sigOps.TransitionSignalTargetFail(targetId, err.Error())
		_, _ = s.sigOps.CheckAndCleanupSignalEvent(tgt.SignalEventID)
		return
	}

	actionCtx := &easyaction.Context{
		Payload: evt.Payload,
	}

	actionErr := engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    sig.ReceiverSpaceID,
		EventType:  "signal",
		ActionName: sig.ReceiverHandler,
		Params: map[string]string{
			"signal_key":          sig.SignalKey,
			"signal_id":           fmt.Sprintf("%d", sig.ID),
			"signal_event_id":     fmt.Sprintf("%d", evt.ID),
			"signal_target_id":    fmt.Sprintf("%d", tgt.ID),
			"emitter_install_id":  fmt.Sprintf("%d", sig.EmitterInstallID),
			"emitter_space_id":    fmt.Sprintf("%d", sig.EmitterSpaceID),
			"receiver_install_id": fmt.Sprintf("%d", sig.ReceiverInstallID),
			"receiver_space_id":   fmt.Sprintf("%d", sig.ReceiverSpaceID),
		},
		Request: actionCtx,
	})

	// Check if action context explicitly called block
	if actionCtx.IsBlocked() {
		reason := actionCtx.BlockReason
		if reason == "" && actionErr != nil {
			reason = actionErr.Error()
		}
		_ = s.sigOps.TransitionSignalTargetBlocked(targetId, reason)
		qq.Println("@SigHub/processSignalTarget: target blocked explicitly by action ctx", targetId, reason)
		return
	}

	if actionErr != nil {
		qq.Println("@SigHub/processSignalTarget/actionErr", actionErr)
		// Check retry policy
		if sig.MaxRetries > 0 && tgt.RetryCount < sig.MaxRetries {
			delayUntil := time.Now().Unix() + sig.RetryDelay
			_ = s.sigOps.TransitionSignalTargetDelay(targetId, delayUntil, tgt.RetryCount+1, actionErr.Error())
			qq.Println("@SigHub/processSignalTarget: scheduled retry", tgt.RetryCount+1, "at", delayUntil)
			return
		}

		_ = s.sigOps.TransitionSignalTargetFail(targetId, actionErr.Error())
		_, _ = s.sigOps.CheckAndCleanupSignalEvent(tgt.SignalEventID)
		return
	}

	_ = s.sigOps.TransitionSignalTargetComplete(targetId)
	_, _ = s.sigOps.CheckAndCleanupSignalEvent(tgt.SignalEventID)
	qq.Println("@SigHub/processSignalTarget: completed", targetId)
}
