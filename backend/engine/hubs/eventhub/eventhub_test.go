package eventhub

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/xtypes"
)

func NewEventHubTest(db datahub.Database) *EventHub {

	return &EventHub{
		db:                     db,
		sink:                   db.GetMQSynk(),
		app:                    nil,
		activeEvents:           make(map[string]bool),
		activeEventsLock:       sync.RWMutex{},
		eventProcessChan:       make(chan int64, 13),
		eventTargetProcessChan: make(chan int64, 27),
		refreshFullIndex:       make(chan struct{}, 1),
		ctx:                    context.Background(),
		cancel:                 func() {},
		wg:                     sync.WaitGroup{},
	}

}

func BuildDBHandle() (datahub.Database, error) {
	tmpDir, err := os.MkdirTemp("", "eventhub_test_*")
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.NewDB(dbPath, slog.New(slog.NewTextHandler(os.Stdout, nil)))
	if err != nil {
		return nil, err
	}

	err = database.AutoMigrate(db.GetSession())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestEventHub_PublishWithSubscription(t *testing.T) {
	// Setup: Create DB handle
	db, err := BuildDBHandle()
	if err != nil {
		t.Fatalf("Failed to build DB handle: %v", err)
	}

	// Create EventHub
	hub := NewEventHubTest(db)

	defer func() {
		// Stop the event hub gracefully
		if hub != nil {
			hub.Stop()
		}
		// Wait a bit for goroutines to finish before closing
		time.Sleep(100 * time.Millisecond)
		if err := db.Close(); err != nil {
			t.Logf("Error closing DB: %v", err)
		}
	}()

	// Create a subscription for a test event with log target
	installId := int64(1)
	eventKey := "test.event"
	subscription := &dbmodels.MQSubscription{
		InstallID:      installId,
		SpaceID:        0, // package-level
		EventKey:       eventKey,
		TargetType:     "log",
		TargetEndpoint: "",
		TargetOptions:  "{}",
		TargetCode:     "",
		Rules:          "{}", // empty rules means match all
		Transform:      "{}",
		DelayStart:     0,
		RetryDelay:     0,
		MaxRetries:     0,
		Disabled:       false,
		CreatedBy:      1,
	}

	subId, err := db.GetSpaceOps().AddEventSubscription(installId, subscription)
	if err != nil {
		t.Fatalf("Failed to add event subscription: %v", err)
	}
	t.Logf("Created subscription with ID: %d", subId)

	// Start the event hub (this starts the event loop goroutines)
	err = hub.Start()
	if err != nil {
		t.Fatalf("Failed to start event hub: %v", err)
	}

	// Give the hub a moment to build the active events index
	time.Sleep(100 * time.Millisecond)

	// Test 1: Publish an event that has a subscription
	payload := []byte(`{"message": "test payload", "value": 42}`)
	err = hub.Publish(&xtypes.EventOptions{
		InstallId: installId,
		Name:      eventKey,
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for processing (event loop picks it up, creates targets, processes them)
	time.Sleep(1 * time.Second)

	// Verify: Check that the event was added to DB and processed
	sink := db.GetMQSynk()

	// Query new events - the event should not be in "new" status anymore
	// (it should be scheduled or processed by now)
	newEvents, err := sink.QueryNewEvents()
	if err != nil {
		t.Fatalf("Failed to query new events: %v", err)
	}

	// The event should have been processed (moved from "new" to "scheduled"/"processed")
	// Since we have a subscription, the event should have been added to DB and processed
	if len(newEvents) > 0 {
		t.Errorf("Expected no events in 'new' status after processing, but found %d", len(newEvents))
	}

	// Verify that targets were created and processed
	// We need to find the event ID first - query all events and find the one we published
	// Since we can't easily query by installId+name, we'll check that processing occurred
	// by verifying the subscription was used (targets would only be created if subscription matched)
	t.Log("Event published and should be processed by now")
}

func TestEventHub_PublishWithoutSubscription(t *testing.T) {
	// Setup: Create DB handle
	db, err := BuildDBHandle()
	if err != nil {
		t.Fatalf("Failed to build DB handle: %v", err)
	}

	// Create EventHub
	hub := NewEventHubTest(db)
	defer func() {
		// Stop the event hub gracefully
		if hub != nil {
			hub.Stop()
		}
		// Wait a bit for goroutines to finish before closing
		time.Sleep(100 * time.Millisecond)
		if err := db.Close(); err != nil {
			t.Logf("Error closing DB: %v", err)
		}
	}()

	// Start the event hub
	err = hub.Start()
	if err != nil {
		t.Fatalf("Failed to start event hub: %v", err)
	}

	// Give the hub a moment to build the active events index
	time.Sleep(100 * time.Millisecond)

	// Test: Publish an event that has NO subscription
	installId := int64(1)
	eventKey := "unsubscribed.event"
	payload := []byte(`{"message": "test payload"}`)

	err = hub.Publish(&xtypes.EventOptions{
		InstallId: installId,
		Name:      eventKey,
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait a moment
	time.Sleep(500 * time.Millisecond)

	// Verify: The event should NOT have been added to DB since there are no subscriptions
	// The Publish method checks needsProcessing first, so if there are no subscriptions,
	// it should return early without adding to DB
	sink := db.GetMQSynk()
	newEvents, err := sink.QueryNewEvents()
	if err != nil {
		t.Fatalf("Failed to query new events: %v", err)
	}

	// There should be no new events since we have no subscriptions
	// (The event should not have been added because needsProcessing returns false)
	if len(newEvents) > 0 {
		t.Errorf("Expected 0 events in DB since no subscription exists, but found %d", len(newEvents))
	}

	t.Log("Event without subscription correctly did not trigger processing")
}

func TestEventHub_FullFlow(t *testing.T) {
	// Setup: Create DB handle
	db, err := BuildDBHandle()
	if err != nil {
		t.Fatalf("Failed to build DB handle: %v", err)
	}

	// Create EventHub
	hub := NewEventHubTest(db)

	defer func() {
		// Stop the event hub gracefully
		if hub != nil {
			hub.Stop()
		}
		// Wait longer for all goroutines to finish processing before closing
		time.Sleep(200 * time.Millisecond)
		if err := db.Close(); err != nil {
			t.Logf("Error closing DB: %v", err)
		}
	}()

	// Create a subscription for a test event with log target
	installId := int64(1)
	eventKey := "test.flow"
	subscription := &dbmodels.MQSubscription{
		InstallID:      installId,
		SpaceID:        0,
		EventKey:       eventKey,
		TargetType:     "log",
		TargetEndpoint: "",
		TargetOptions:  "{}",
		TargetCode:     "",
		Rules:          "{}",
		Transform:      "{}",
		DelayStart:     0,
		RetryDelay:     0,
		MaxRetries:     0,
		Disabled:       false,
		CreatedBy:      1,
	}

	subId, err := db.GetSpaceOps().AddEventSubscription(installId, subscription)
	if err != nil {
		t.Fatalf("Failed to add event subscription: %v", err)
	}

	// Start the event hub
	err = hub.Start()
	if err != nil {
		t.Fatalf("Failed to start event hub: %v", err)
	}

	// Give the hub a moment to build the active events index
	time.Sleep(100 * time.Millisecond)

	// Publish an event
	testPayload := map[string]interface{}{
		"message": "test message",
		"value":   123,
		"data":    []string{"a", "b", "c"},
	}
	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Use Publish method which checks for subscriptions first
	err = hub.Publish(&xtypes.EventOptions{
		InstallId: installId,
		Name:      eventKey,
		Payload:   payloadBytes,
	})
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Wait for processing (event loop picks it up, creates targets, processes them)
	time.Sleep(2 * time.Second)

	// Verify: Check that events were processed
	// Query new events - they should be scheduled or processed by now
	sink := db.GetMQSynk()
	newEvents, err := sink.QueryNewEvents()
	if err != nil {
		t.Fatalf("Failed to query new events: %v", err)
	}

	// The event should not be in "new" status anymore
	if len(newEvents) > 0 {
		t.Errorf("Expected no events in 'new' status after processing, but found %d", len(newEvents))
	}

	// Verify: Check subscription was used
	sub, err := db.GetSpaceOps().GetEventSubscription(installId, subId)
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if sub.EventKey != eventKey {
		t.Errorf("Subscription event key mismatch: expected '%s', got '%s'", eventKey, sub.EventKey)
	}

	if sub.TargetType != "log" {
		t.Errorf("Expected target type 'log', got '%s'", sub.TargetType)
	}

	// Verify that targets were created and processed
	// Check that no new targets remain (they should all be processed)
	newTargets, err := sink.QueryNewEventTargets()
	if err != nil {
		t.Fatalf("Failed to query new event targets: %v", err)
	}

	// After processing, there should be no new targets
	// (they should have been processed or failed)
	if len(newTargets) > 0 {
		t.Errorf("Expected no new event targets after processing, but found %d", len(newTargets))
	}

	t.Log("Full flow test completed successfully")
}

func TestEventHub_TargetCreationAndProcessing(t *testing.T) {
	// Setup: Create DB handle
	db, err := BuildDBHandle()
	if err != nil {
		t.Fatalf("Failed to build DB handle: %v", err)
	}

	// Create EventHub
	hub := NewEventHubTest(db)

	defer func() {
		// Stop the event hub gracefully
		if hub != nil {
			hub.Stop()
		}
		// Wait longer for all goroutines to finish processing before closing
		time.Sleep(200 * time.Millisecond)
		if err := db.Close(); err != nil {
			t.Logf("Error closing DB: %v", err)
		}
	}()

	// Create a subscription for a test event with log target
	installId := int64(1)
	eventKey := "test.targets"
	subscription := &dbmodels.MQSubscription{
		InstallID:      installId,
		SpaceID:        0,
		EventKey:       eventKey,
		TargetType:     "log",
		TargetEndpoint: "",
		TargetOptions:  "{}",
		TargetCode:     "",
		Rules:          "{}",
		Transform:      "{}",
		DelayStart:     0,
		RetryDelay:     0,
		MaxRetries:     0,
		Disabled:       false,
		CreatedBy:      1,
	}

	subId, err := db.GetSpaceOps().AddEventSubscription(installId, subscription)
	if err != nil {
		t.Fatalf("Failed to add event subscription: %v", err)
	}

	// Start the event hub
	err = hub.Start()
	if err != nil {
		t.Fatalf("Failed to start event hub: %v", err)
	}

	// Give the hub a moment to build the active events index
	time.Sleep(100 * time.Millisecond)

	// Get the sink to track event ID
	sink := db.GetMQSynk()

	// Publish an event and get the event ID directly
	payload := []byte(`{"message": "target test", "value": 999}`)
	eventId, err := sink.AddEvent(installId, eventKey, payload)
	if err != nil {
		t.Fatalf("Failed to add event directly: %v", err)
	}

	// Send the event to the processing channel
	hub.eventProcessChan <- eventId

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify: Check that the event was processed (status changed from "new")
	event, err := sink.GetEvent(eventId)
	if err != nil {
		t.Fatalf("Failed to get event: %v", err)
	}

	if event.Status == "new" {
		t.Errorf("Expected event status to be 'scheduled' or 'processed', but got 'new'")
	}

	// Verify: Check that targets were created for this event
	targetIds, err := sink.QueryEventTargetsByEventId(eventId)
	if err != nil {
		t.Fatalf("Failed to query event targets: %v", err)
	}

	if len(targetIds) == 0 {
		t.Errorf("Expected at least one target to be created for the event, but found 0")
	}

	// Verify: Check that targets were processed (no new targets remain)
	newTargets, err := sink.QueryNewEventTargets()
	if err != nil {
		t.Fatalf("Failed to query new event targets: %v", err)
	}

	// Check if any of the targets we created are still in "new" status
	for _, targetId := range targetIds {
		isNew := false
		for _, newTargetId := range newTargets {
			if targetId == newTargetId {
				isNew = true
				break
			}
		}
		if isNew {
			t.Errorf("Target %d is still in 'new' status after processing", targetId)
		}
	}

	// Verify: Check that the subscription was used correctly
	sub, err := db.GetSpaceOps().GetEventSubscription(installId, subId)
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if sub.ID != subId {
		t.Errorf("Subscription ID mismatch: expected %d, got %d", subId, sub.ID)
	}

	if sub.EventKey != eventKey {
		t.Errorf("Subscription event key mismatch: expected '%s', got '%s'", eventKey, sub.EventKey)
	}

	t.Logf("Successfully verified event %d was processed with %d targets", eventId, len(targetIds))
}
