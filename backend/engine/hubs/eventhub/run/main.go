package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/eslayer"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/k0kubun/pp"
)

// this is a testing harness for eventhub/eslayer (meaning Event slayer) (kind a like job processering serveice)
// eslayer should be robust, handle various scenarios

func Show(a ...any) (n int, err error) {

	return pp.Println(a...)
}

type TestResults struct {
	mu                sync.Mutex
	successCount      int
	failCount         int
	retryCount        int
	delayedCount      int
	handlerNotFound   int
	totalProcessed    int
	expectedProcessed int
}

func (tr *TestResults) recordSuccess() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.successCount++
	tr.totalProcessed++
}

func (tr *TestResults) recordFail() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.failCount++
}

func (tr *TestResults) recordRetry() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.retryCount++
}

func (tr *TestResults) recordDelayed() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.delayedCount++
}

func (tr *TestResults) recordHandlerNotFound() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.handlerNotFound++
}

func (tr *TestResults) print() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	Show("=== Test Results ===")
	Show("Total Processed:", tr.totalProcessed, "/", tr.expectedProcessed)
	Show("Success Count:", tr.successCount)
	Show("Fail Count:", tr.failCount)
	Show("Retry Count:", tr.retryCount)
	Show("Delayed Count:", tr.delayedCount)
	Show("Handler Not Found:", tr.handlerNotFound)
	Show("===================")
}

func main() {

	qq.Enabled = true

	tmpFolder, err := os.MkdirTemp("", "estest_*")
	if err != nil {
		Show("Failed to create temp directory", "err", err)
		return
	}
	Show("tmpFolder", tmpFolder)

	defer os.RemoveAll(tmpFolder)

	sdb, err := sql.Open("sqlite3", fmt.Sprintf("file:%s/main1.db?mode=rwc&_journal_mode=WAL&_busy_timeout=1000", tmpFolder))
	if err != nil {
		log.Fatal(err)
	}
	defer sdb.Close()

	sdb.SetMaxOpenConns(1)

	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := database.FromSqlHandle(sdb)
	if err != nil {
		log.Fatal(err)
	}

	results := &TestResults{
		expectedProcessed: 5, // We expect 5 successful handlers to run
	}

	failCountMap := make(map[string]int)
	var failMu sync.Mutex

	handlers := map[string]evtype.Handler{
		"test": func(ex *evtype.TExecution) error {
			Show("==> Handler 'test' called for event:", ex.Subscription.EventKey)

			switch ex.Subscription.EventKey {
			case "test-success":
				Show("✓ test-success: Processing successful event")
				results.recordSuccess()
				return nil

			case "test-fail":
				Show("✗ test-fail: Simulating permanent failure")
				results.recordFail()
				return errors.New("permanent failure - not retryable")

			case "test-retry-then-success":
				failMu.Lock()
				count := failCountMap["test-retry-then-success"]
				failCountMap["test-retry-then-success"] = count + 1
				failMu.Unlock()

				if count < 2 {
					Show("⟳ test-retry-then-success: Retry attempt", count+1, "- failing")
					results.recordRetry()
					ex.RetryAble = true
					return errors.New("temporary failure - will retry")
				}

				Show("✓ test-retry-then-success: Success after", count, "retries")
				results.recordSuccess()
				return nil

			case "test-delayed":
				Show("⏱ test-delayed: Processing delayed event")
				results.recordDelayed()
				results.recordSuccess()
				return nil

			case "test-multiple-subs":
				Show("✓ test-multiple-subs: Processing event with multiple subscriptions")
				results.recordSuccess()
				return nil

			default:
				Show("? Unknown event key:", ex.Subscription.EventKey)
				return nil
			}
		},
		"webhook": func(ex *evtype.TExecution) error {
			Show("==> Handler 'webhook' called for event:", ex.Subscription.EventKey)
			results.recordSuccess()
			return nil
		},
	}

	// Create event subscriptions
	Show("\n=== Setting up Event Subscriptions ===")
	installID := int64(1)
	spaceID := int64(1)

	subscriptions := []dbmodels.MQSubscription{
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-success",
			TargetType:     "test",
			TargetEndpoint: "",
			Rules:          "{}",
			CreatedBy:      1,
			Disabled:       false,
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-fail",
			TargetType:     "test",
			TargetEndpoint: "",
			Rules:          "{}",
			CreatedBy:      1,
			Disabled:       false,
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-retry-then-success",
			TargetType:     "test",
			TargetEndpoint: "",
			Rules:          "{}",
			RetryDelay:     1, // 1 second retry delay
			MaxRetries:     5, // Allow up to 5 retries
			CreatedBy:      1,
			Disabled:       false,
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-delayed",
			TargetType:     "test",
			TargetEndpoint: "",
			Rules:          "{}",
			DelayStart:     2, // 2 second delay before processing
			CreatedBy:      1,
			Disabled:       false,
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-multiple-subs",
			TargetType:     "test",
			TargetEndpoint: "",
			Rules:          "{}",
			CreatedBy:      1,
			Disabled:       false,
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-multiple-subs",
			TargetType:     "webhook",
			TargetEndpoint: "http://example.com/webhook",
			Rules:          "{}",
			CreatedBy:      1,
			Disabled:       false,
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-disabled",
			TargetType:     "test",
			TargetEndpoint: "",
			Rules:          "{}",
			CreatedBy:      1,
			Disabled:       true, // This should not be processed
		},
		{
			InstallID:      installID,
			SpaceID:        spaceID,
			EventKey:       "test-handler-not-found",
			TargetType:     "nonexistent",
			TargetEndpoint: "",
			Rules:          "{}",
			CreatedBy:      1,
			Disabled:       false,
		},
	}

	for i, sub := range subscriptions {
		id, err := db.GetSpaceOps().AddEventSubscription(installID, &sub)
		if err != nil {
			Show(fmt.Sprintf("Failed to create subscription %d", i), "err", err)
			return
		}
		Show(fmt.Sprintf("✓ Created subscription %d: %s -> %s (ID: %d)", i+1, sub.EventKey, sub.TargetType, id))
	}

	// Start the event layer
	Show("\n=== Starting ESLayer ===")
	eslayer := eslayer.NewESLayer(db, handlers)

	err = eslayer.Start()
	if err != nil {
		Show("Failed to start eslayer", "err", err)
		return
	}

	// Give the system a moment to initialize
	time.Sleep(2 * time.Second)

	// Add test events
	Show("\n=== Adding Test Events ===")

	events := []struct {
		name    string
		payload string
	}{
		{"test-success", `{"message": "This should succeed"}`},
		{"test-fail", `{"message": "This should fail"}`},
		{"test-retry-then-success", `{"message": "This should retry then succeed"}`},
		{"test-delayed", `{"message": "This should be delayed"}`},
		{"test-multiple-subs", `{"message": "This should trigger multiple subscriptions"}`},
		{"test-disabled", `{"message": "This should not be processed"}`},
		{"test-handler-not-found", `{"message": "This should fail - handler not found"}`},
	}

	for _, evt := range events {
		eventID, err := db.GetMQSynk().AddEvent(installID, evt.name, []byte(evt.payload))
		if err != nil {
			Show("Failed to add event", evt.name, "err", err)
		} else {
			Show(fmt.Sprintf("✓ Added event: %s (ID: %d)", evt.name, eventID))
			eslayer.NotifyNewEvent(eventID)
		}
	}

	// Wait for processing
	Show("\n=== Processing Events (waiting 35 seconds) ===")
	Show("This allows time for:")
	Show("  - Immediate event processing")
	Show("  - Retry attempts (with delays)")
	Show("  - Delayed event execution")
	Show("  - Fallback timer triggers")

	time.Sleep(35 * time.Second)

	// Stop the event layer
	Show("\n=== Stopping ESLayer ===")
	eslayer.Stop()

	// Print results
	Show("\n")
	results.print()

	// Verify expectations
	Show("\n=== Verification ===")
	results.mu.Lock()
	defer results.mu.Unlock()

	if results.totalProcessed >= results.expectedProcessed {
		Show("✓ Test PASSED: All expected events were processed")
	} else {
		Show("✗ Test FAILED: Not all events were processed")
	}

	if results.retryCount >= 2 {
		Show("✓ Retry mechanism working (", results.retryCount, "retries observed)")
	} else {
		Show("⚠ Retry mechanism may not be working properly")
	}

	if results.delayedCount > 0 {
		Show("✓ Delayed execution working")
	} else {
		Show("⚠ Delayed execution may not be working properly")
	}

	Show("\n=== Test Complete ===")
}
