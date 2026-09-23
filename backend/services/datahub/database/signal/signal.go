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

func (s *SignalOperations) AddSignalEvent(signalId int64, payload []byte, metadata map[string]any) (int64, error) {
	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		metaBytes = []byte("{}")
	}

	now := time.Now()
	event := &dbmodels.SignalEvent{
		SignalID:     signalId,
		Payload:      payload,
		Metadata:     string(metaBytes),
		Status:       "new",
		DelayedUntil: 0,
		RetryCount:   0,
		CreatedAt:    &now,
		UpdatedAt:    &now,
		ExtraMeta:    "{}",
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

func (s *SignalOperations) UpdateSignalEvent(id int64, data map[string]any) error {
	data["updated_at"] = time.Now()
	return s.eventTable().Find(db.Cond{"id": id}).Update(data)
}

func (s *SignalOperations) QuerySignalEvents(installId int64, signalId int64, status string, limit, offset int64) ([]dbmodels.SignalEvent, error) {
	cond := db.Cond{}
	if signalId > 0 {
		cond["signal_id"] = signalId
	} else if installId > 0 {
		// Find signals related to this install
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
			return []dbmodels.SignalEvent{}, nil
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

	var events []dbmodels.SignalEvent
	err := s.eventTable().Find(cond).OrderBy("-id").Limit(int(limit)).Offset(int(offset)).All(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (s *SignalOperations) QueryNewSignalEvents() ([]int64, error) {
	events := make([]struct {
		ID int64 `db:"id"`
	}, 0)

	err := s.eventTable().Find(db.Cond{"status": "new"}).Select("id").All(&events)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	return ids, nil
}

func (s *SignalOperations) QueryDelayExpiredSignalEvents() ([]int64, error) {
	now := time.Now().Unix()
	events := make([]struct {
		ID int64 `db:"id"`
	}, 0)

	err := s.eventTable().Find(db.Cond{
		"status":           "delayed",
		"delayed_until <=": now,
	}).Select("id").All(&events)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	return ids, nil
}

// State transitions

func (s *SignalOperations) TransitionSignalEventStart(id int64) (*dbmodels.SignalEvent, error) {
	err := s.UpdateSignalEvent(id, map[string]any{
		"status": "processing",
	})
	if err != nil {
		return nil, err
	}
	return s.GetSignalEvent(id)
}

func (s *SignalOperations) TransitionSignalEventComplete(id int64) error {
	return s.UpdateSignalEvent(id, map[string]any{
		"status": "processed",
	})
}

func (s *SignalOperations) TransitionSignalEventDelay(id int64, delayUntil int64, retryCount int64, errorMsg string) error {
	return s.UpdateSignalEvent(id, map[string]any{
		"status":          "delayed",
		"delayed_until":   delayUntil,
		"retry_count":     retryCount,
		"last_retried_at": time.Now().Unix(),
		"error":           errorMsg,
	})
}

func (s *SignalOperations) TransitionSignalEventFail(id int64, errorMsg string) error {
	return s.UpdateSignalEvent(id, map[string]any{
		"status": "failed",
		"error":  errorMsg,
	})
}

func (s *SignalOperations) TransitionSignalEventExpired(id int64) error {
	return s.UpdateSignalEvent(id, map[string]any{
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
