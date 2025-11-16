package event

import (
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type EventOperations struct {
	db db.Session
}

func NewEventOperations(db db.Session) *EventOperations {
	return &EventOperations{
		db: db,
	}
}

func (d *EventOperations) AddEvent(installId int64, name string, payload []byte) (int64, error) {
	event := &dbmodels.MQEvent{
		InstallID: installId,
		Name:      name,
		Payload:   payload,
		Status:    "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	r, err := d.eventTable().Insert(event)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *EventOperations) GetEvent(id int64) (*dbmodels.MQEvent, error) {
	event := &dbmodels.MQEvent{}
	err := d.eventTable().Find(db.Cond{"id": id}).One(event)
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (d *EventOperations) UpdateEvent(id int64, data map[string]any) error {
	data["updated_at"] = time.Now()
	return d.eventTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *EventOperations) QueryNewEvents() ([]int64, error) {
	events := make([]struct {
		ID int64 `db:"id"`
	}, 0)

	err := d.eventTable().Find(db.Cond{"status": "new"}).Select("id").All(&events)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(events))
	for i, e := range events {
		ids[i] = e.ID
	}
	return ids, nil
}

func (d *EventOperations) CreateEventTargets(eventId int64) ([]int64, error) {
	// Get the event to find matching subscriptions
	event, err := d.GetEvent(eventId)
	if err != nil {
		return nil, err
	}

	// Query subscriptions that match this event
	subscriptions := make([]dbmodels.EventSubscriptionLite, 0)
	err = d.subscriptionTable().Find(db.Cond{
		"install_id": event.InstallID,
		"event_key":  event.Name,
		"disabled":   false,
	}).All(&subscriptions)
	if err != nil {
		return nil, err
	}

	// Create event targets for each matching subscription
	targetIds := make([]int64, 0, len(subscriptions))
	for _, sub := range subscriptions {
		target := &dbmodels.MQEventTarget{
			EventID:        eventId,
			SubscriptionID: sub.ID,
			Status:         "new",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		r, err := d.eventTargetTable().Insert(target)
		if err != nil {
			return nil, err
		}

		targetIds = append(targetIds, r.ID().(int64))
	}

	return targetIds, nil
}

func (d *EventOperations) QueryNewEventTargets() ([]int64, error) {
	targets := make([]struct {
		ID int64 `db:"id"`
	}, 0)

	err := d.eventTargetTable().Find(db.Cond{"status": "new"}).Select("id").All(&targets)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(targets))
	for i, t := range targets {
		ids[i] = t.ID
	}
	return ids, nil
}

func (d *EventOperations) QueryDelayExpiredTargets() ([]int64, error) {
	now := time.Now().Unix()
	entityIds := make([]dbmodels.EntityId, 0)
	err := d.eventTargetTable().Find(db.Cond{
		"status":           "delayed",
		"delayed_until <=": now,
	}).Select("id").All(&entityIds)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(entityIds))
	for i, e := range entityIds {
		ids[i] = e.Id
	}

	return ids, nil
}

func (d *EventOperations) QueryEventTargetsByEventId(eventId int64) ([]int64, error) {
	entityIds := make([]dbmodels.EntityId, 0)

	err := d.eventTargetTable().Find(db.Cond{"event_id": eventId}).Select("id").All(&entityIds)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(entityIds))
	for i, e := range entityIds {
		ids[i] = e.Id
	}
	return ids, nil
}

func (d *EventOperations) UpdateEventTarget(id int64, data map[string]any) error {
	data["updated_at"] = time.Now()
	return d.eventTargetTable().Find(db.Cond{"id": id}).Update(data)
}

// transition

func (d *EventOperations) TransitionTargetStart(id int64) (*dbmodels.MQEventTarget, error) {
	// Update status to "processing" and return the target
	err := d.UpdateEventTarget(id, map[string]any{
		"status": "processing",
	})
	if err != nil {
		return nil, err
	}

	target := &dbmodels.MQEventTarget{}
	err = d.eventTargetTable().Find(db.Cond{"id": id}).One(target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (d *EventOperations) TransitionTargetComplete(evtId, targetId int64) error {
	err := d.UpdateEventTarget(targetId, map[string]any{
		"status":   "processed",
		"event_id": evtId,
	})

	if err != nil {
		return err
	}

	count, err := d.eventTargetTable().Find(db.Cond{
		"event_id":  evtId,
		"status !=": "processed",
	}).Count()

	if count == 0 {
		return d.UpdateEvent(evtId, map[string]any{
			"status": "processed",
		})
	}

	if err != nil {
		return err
	}

	return nil
}

func (d *EventOperations) TransitionTargetFail(evtId, targetId int64, errorMsg string) error {
	return d.UpdateEventTarget(targetId, map[string]any{
		"status": "failed",
		"error":  errorMsg,
	})
}

// Private helper methods

func (d *EventOperations) eventTable() db.Collection {
	return d.db.Collection("MQEvents")
}

func (d *EventOperations) eventTargetTable() db.Collection {
	return d.db.Collection("MQEventTargets")
}

func (d *EventOperations) subscriptionTable() db.Collection {
	return d.db.Collection("EventSubscriptions")
}
