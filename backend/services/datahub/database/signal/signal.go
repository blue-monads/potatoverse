package signal

import (
	"encoding/json"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type SignalOperations struct {
	db db.Session
}

func NewSignalOperations(db db.Session) *SignalOperations {
	return &SignalOperations{
		db: db,
	}
}

// Signals operations

func (s *SignalOperations) AddSignal(data *dbmodels.Signal) (int64, error) {
	now := time.Now()
	data.CreatedAt = &now
	data.UpdatedAt = &now

	r, err := s.signalTable().Insert(data)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (s *SignalOperations) GetSignal(id int64) (*dbmodels.Signal, error) {
	sig := &dbmodels.Signal{}
	err := s.signalTable().Find(db.Cond{"id": id}).One(sig)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignalOperations) UpdateSignal(id int64, data map[string]any) error {
	data["updated_at"] = time.Now()
	return s.signalTable().Find(db.Cond{"id": id}).Update(data)
}

func (s *SignalOperations) RemoveSignal(id int64) error {
	return s.signalTable().Find(db.Cond{"id": id}).Delete()
}

func (s *SignalOperations) QuerySignals(installId int64, spaceId int64, cond map[any]any) ([]dbmodels.Signal, error) {
	queryCond := db.Cond{}
	for k, v := range cond {
		queryCond[k] = v
	}

	var orExpr any

	// Filter by installId if provided (> 0)
	if installId > 0 {
		role, _ := queryCond["role"].(string)
		delete(queryCond, "role")

		switch role {
		case "emitter":
			queryCond["emitter_install_id"] = installId
			if spaceId > 0 {
				queryCond["emitter_space_id"] = spaceId
			}
		case "receiver":
			queryCond["receiver_install_id"] = installId
			if spaceId > 0 {
				queryCond["receiver_space_id"] = spaceId
			}
		default:
			// Match either emitter or receiver
			if spaceId > 0 {
				orExpr = db.Or(
					db.Cond{"emitter_install_id": installId, "emitter_space_id": spaceId},
					db.Cond{"receiver_install_id": installId, "receiver_space_id": spaceId},
				)
			} else {
				orExpr = db.Or(
					db.Cond{"emitter_install_id": installId},
					db.Cond{"receiver_install_id": installId},
				)
			}
		}
	}

	var terms []any
	if len(queryCond) > 0 {
		terms = append(terms, queryCond)
	}
	if orExpr != nil {
		terms = append(terms, orExpr)
	}

	var signals []dbmodels.Signal
	err := s.signalTable().Find(terms...).OrderBy("-id").All(&signals)
	if err != nil {
		return nil, err
	}
	return signals, nil
}

func (s *SignalOperations) QueryAllActiveSignals() ([]dbmodels.SignalLite, error) {
	var signals []dbmodels.SignalLite
	err := s.signalTable().Find(db.Cond{"disabled": false}).All(&signals)
	if err != nil {
		return nil, err
	}
	return signals, nil
}

func (s *SignalOperations) GetSignalsForEmitter(emitterInstallId, emitterSpaceId int64, signalKey string) ([]dbmodels.SignalLite, error) {
	cond := db.Cond{
		"signal_key":         signalKey,
		"emitter_install_id": emitterInstallId,
		"disabled":           false,
	}

	var all []dbmodels.SignalLite
	err := s.signalTable().Find(cond).All(&all)
	if err != nil {
		return nil, err
	}

	// Filter by emitterSpaceId if space-specific
	res := make([]dbmodels.SignalLite, 0, len(all))
	for _, sig := range all {
		if sig.EmitterSpaceID == 0 || sig.EmitterSpaceID == emitterSpaceId {
			res = append(res, sig)
		}
	}
	return res, nil
}

// Signal Events operations

func (s *SignalOperations) AddSignalEvent(signalEventKey string, payload []byte, metadata map[string]any) (int64, error) {
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		metaBytes = []byte("{}")
	}

	now := time.Now()
	event := &dbmodels.SignalEvent{
		SignalEventKey: signalEventKey,
		Payload:        payload,
		Metadata:       string(metaBytes),
		CreatedAt:      &now,
		ExtraMeta:      "{}",
	}

	r, err := s.eventTable().Insert(event)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (s *SignalOperations) GetSignalEvent(id int64) (*dbmodels.SignalEvent, error) {
	evt := &dbmodels.SignalEvent{}
	err := s.eventTable().Find(db.Cond{"id": id}).One(evt)
	if err != nil {
		return nil, err
	}
	return evt, nil
}

func (s *SignalOperations) DeleteSignalEvent(id int64) error {
	return s.eventTable().Find(db.Cond{"id": id}).Delete()
}

// Signal Targets operations

func (s *SignalOperations) AddSignalTarget(signalEventId, signalId int64, metadata map[string]any) (int64, error) {
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		metaBytes = []byte("{}")
	}

	target := &dbmodels.SignalTarget{
		SignalEventID: signalEventId,
		SignalID:      signalId,
		Status:        "new",
		Metadata:      string(metaBytes),
		DelayedUntil:  0,
		RetryCount:    0,
		LastRetriedAt: 0,
		Error:         "",
		ExtraMeta:     "{}",
	}

	r, err := s.targetTable().Insert(target)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (s *SignalOperations) GetSignalTarget(id int64) (*dbmodels.SignalTarget, error) {
	tgt := &dbmodels.SignalTarget{}
	err := s.targetTable().Find(db.Cond{"id": id}).One(tgt)
	if err != nil {
		return nil, err
	}
	return tgt, nil
}

func (s *SignalOperations) UpdateSignalTarget(id int64, data map[string]any) error {
	return s.targetTable().Find(db.Cond{"id": id}).Update(data)
}

func (s *SignalOperations) QuerySignalTargets(installId int64, signalId int64, status string, limit, offset int64) ([]dbmodels.SignalTargetWithDetails, error) {
	cond := db.Cond{}
	if signalId > 0 {
		cond["signal_id"] = signalId
	} else if installId > 0 {
		var signalIds []struct {
			ID int64 `db:"id"`
		}
		err := s.signalTable().Find(db.Or(
			db.Cond{"emitter_install_id": installId},
			db.Cond{"receiver_install_id": installId},
		)).Select("id").All(&signalIds)
		if err != nil {
			return nil, err
		}
		if len(signalIds) == 0 {
			return []dbmodels.SignalTargetWithDetails{}, nil
		}
		ids := make([]int64, len(signalIds))
		for i, s := range signalIds {
			ids[i] = s.ID
		}
		cond["signal_id IN"] = ids
	}

	if status != "" {
		cond["status"] = status
	}

	if limit <= 0 {
		limit = 100
	}

	var targets []dbmodels.SignalTarget
	err := s.targetTable().Find(cond).OrderBy("-id").Limit(int(limit)).Offset(int(offset)).All(&targets)
	if err != nil {
		return nil, err
	}

	results := make([]dbmodels.SignalTargetWithDetails, 0, len(targets))
	sigCache := make(map[int64]*dbmodels.Signal)
	evtCache := make(map[int64]*dbmodels.SignalEvent)

	for _, tgt := range targets {
		item := dbmodels.SignalTargetWithDetails{
			SignalTarget: tgt,
		}

		// Fetch Signal info
		sig, ok := sigCache[tgt.SignalID]
		if !ok {
			sig, _ = s.GetSignal(tgt.SignalID)
			sigCache[tgt.SignalID] = sig
		}
		if sig != nil {
			item.SignalKey = sig.SignalKey
			item.EmitterInstallID = sig.EmitterInstallID
			item.EmitterSpaceID = sig.EmitterSpaceID
			item.ReceiverInstallID = sig.ReceiverInstallID
			item.ReceiverSpaceID = sig.ReceiverSpaceID
			item.ReceiverHandler = sig.ReceiverHandler
		}

		// Fetch Event payload if still present
		evt, ok := evtCache[tgt.SignalEventID]
		if !ok {
			evt, _ = s.GetSignalEvent(tgt.SignalEventID)
			evtCache[tgt.SignalEventID] = evt
		}
		if evt != nil {
			item.Payload = evt.Payload
			item.CreatedAt = evt.CreatedAt
		}

		results = append(results, item)
	}

	return results, nil
}

func (s *SignalOperations) QueryNewSignalTargets() ([]int64, error) {
	targets := make([]struct {
		ID int64 `db:"id"`
	}, 0)

	err := s.targetTable().Find(db.Cond{"status": "new"}).Select("id").All(&targets)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(targets))
	for i, t := range targets {
		ids[i] = t.ID
	}
	return ids, nil
}

func (s *SignalOperations) QueryDelayExpiredSignalTargets() ([]int64, error) {
	now := time.Now().Unix()
	targets := make([]struct {
		ID int64 `db:"id"`
	}, 0)

	err := s.targetTable().Find(db.Cond{
		"status":           "delayed",
		"delayed_until <=": now,
	}).Select("id").All(&targets)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(targets))
	for i, t := range targets {
		ids[i] = t.ID
	}
	return ids, nil
}

// CheckAndCleanupSignalEvent checks if all targets for a signal_event_id have reached a final status.
// Final statuses: "processed", "failed", "expired".
// If all targets are in a final status, deletes the SignalEvents row (releasing payload).
func (s *SignalOperations) CheckAndCleanupSignalEvent(signalEventId int64) (bool, error) {
	if signalEventId <= 0 {
		return false, nil
	}

	var targets []struct {
		Status string `db:"status"`
	}
	err := s.targetTable().Find(db.Cond{"signal_event_id": signalEventId}).Select("status").All(&targets)
	if err != nil {
		return false, err
	}

	if len(targets) == 0 {
		_ = s.DeleteSignalEvent(signalEventId)
		return true, nil
	}

	for _, t := range targets {
		switch t.Status {
		case "processed", "failed", "expired":
			// final status
		default:
			// still active: "new", "scheduled", "delayed", "blocked"
			return false, nil
		}
	}

	// All targets are in a final state, delete event to free payload
	err = s.DeleteSignalEvent(signalEventId)
	if err != nil {
		return false, err
	}
	return true, nil
}

// State transitions

func (s *SignalOperations) TransitionSignalTargetStart(id int64) (*dbmodels.SignalTarget, error) {
	err := s.UpdateSignalTarget(id, map[string]any{
		"status": "scheduled",
	})
	if err != nil {
		return nil, err
	}
	return s.GetSignalTarget(id)
}

func (s *SignalOperations) TransitionSignalTargetComplete(id int64) error {
	return s.UpdateSignalTarget(id, map[string]any{
		"status": "processed",
	})
}

func (s *SignalOperations) TransitionSignalTargetDelay(id int64, delayUntil int64, retryCount int64, errorMsg string) error {
	return s.UpdateSignalTarget(id, map[string]any{
		"status":          "delayed",
		"delayed_until":   delayUntil,
		"retry_count":     retryCount,
		"last_retried_at": time.Now().Unix(),
		"error":           errorMsg,
	})
}

func (s *SignalOperations) TransitionSignalTargetBlocked(id int64, reason string) error {
	return s.UpdateSignalTarget(id, map[string]any{
		"status": "blocked",
		"error":  reason,
	})
}

func (s *SignalOperations) TransitionSignalTargetUnblock(id int64) error {
	return s.UpdateSignalTarget(id, map[string]any{
		"status": "new",
		"error":  "",
	})
}

func (s *SignalOperations) TransitionSignalTargetFail(id int64, errorMsg string) error {
	return s.UpdateSignalTarget(id, map[string]any{
		"status": "failed",
		"error":  errorMsg,
	})
}

func (s *SignalOperations) TransitionSignalTargetExpired(id int64) error {
	return s.UpdateSignalTarget(id, map[string]any{
		"status": "expired",
	})
}

// Table references

func (s *SignalOperations) signalTable() db.Collection {
	return s.db.Collection("Signals")
}

func (s *SignalOperations) eventTable() db.Collection {
	return s.db.Collection("SignalEvents")
}

func (s *SignalOperations) targetTable() db.Collection {
	return s.db.Collection("SignalTargets")
}
