package signal

import (
	"database/sql"
	"testing"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/datahub/provider/sqlitecore"
	"github.com/upper/db/v4/adapter/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func TestQuerySignals(t *testing.T) {
	sdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sdb.Close()

	schema := sqlitecore.Get()
	_, err = sdb.Exec(schema)
	if err != nil {
		t.Fatal(err)
	}

	sess, err := sqlite.New(sdb)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.Close()

	ops := NewSignalOperations(sess)

	// Test QuerySignals with space_id = 1
	signals, err := ops.QuerySignals(1, 1, nil)
	if err != nil {
		t.Fatalf("QuerySignals failed: %v", err)
	}
	t.Logf("Signals: %v", signals)

	// Test QuerySignals with space_id = 0
	signals, err = ops.QuerySignals(1, 0, nil)
	if err != nil {
		t.Fatalf("QuerySignals space_id=0 failed: %v", err)
	}

	// Test QuerySignals with role = emitter
	signals, err = ops.QuerySignals(1, 1, map[any]any{"role": "emitter"})
	if err != nil {
		t.Fatalf("QuerySignals role=emitter failed: %v", err)
	}

	// Test QuerySignals with role = receiver
	signals, err = ops.QuerySignals(1, 1, map[any]any{"role": "receiver"})
	if err != nil {
		t.Fatalf("QuerySignals role=receiver failed: %v", err)
	}

	// Test AddSignal and then Query
	sigId, err := ops.AddSignal(&dbmodels.Signal{
		SignalKey:         "test_signal",
		EmitterInstallID:  1,
		EmitterSpaceID:    1,
		ReceiverInstallID: 2,
		ReceiverSpaceID:   2,
		ReceiverHandler:   "handle_test",
		ManagedBy:         "both",
	})
	if err != nil {
		t.Fatalf("AddSignal failed: %v", err)
	}

	signals, err = ops.QuerySignals(1, 1, nil)
	if err != nil {
		t.Fatalf("QuerySignals after insert failed: %v", err)
	}
	if len(signals) != 1 || signals[0].ID != sigId {
		t.Fatalf("Expected 1 signal with ID %d, got %v", sigId, signals)
	}

	// Test SignalEvents and SignalTargets with shared payload
	eventId, err := ops.AddSignalEvent("test_signal", []byte("shared_blob_payload"), map[string]any{"source": "test"})
	if err != nil {
		t.Fatalf("AddSignalEvent failed: %v", err)
	}

	targetId1, err := ops.AddSignalTarget(eventId, sigId, nil)
	if err != nil {
		t.Fatalf("AddSignalTarget 1 failed: %v", err)
	}

	targetId2, err := ops.AddSignalTarget(eventId, sigId, nil)
	if err != nil {
		t.Fatalf("AddSignalTarget 2 failed: %v", err)
	}

	// Test QuerySignalTargets
	targets, err := ops.QuerySignalTargets(1, sigId, "", 10, 0)
	if err != nil {
		t.Fatalf("QuerySignalTargets failed: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(targets))
	}
	if string(targets[0].Payload) != "shared_blob_payload" {
		t.Fatalf("Expected shared payload 'shared_blob_payload', got '%s'", string(targets[0].Payload))
	}
	if targets[0].SignalKey != "test_signal" {
		t.Fatalf("Expected signal_key 'test_signal', got '%s'", targets[0].SignalKey)
	}

	// Test Transition to blocked
	err = ops.TransitionSignalTargetBlocked(targetId1, "app waiting for dependency")
	if err != nil {
		t.Fatalf("TransitionSignalTargetBlocked failed: %v", err)
	}
	tgt1, _ := ops.GetSignalTarget(targetId1)
	if tgt1.Status != "blocked" || tgt1.Error != "app waiting for dependency" {
		t.Fatalf("Expected target 1 to be blocked, got status=%s error=%s", tgt1.Status, tgt1.Error)
	}

	// Test Transition to unblock (new)
	err = ops.TransitionSignalTargetUnblock(targetId1)
	if err != nil {
		t.Fatalf("TransitionSignalTargetUnblock failed: %v", err)
	}
	tgt1, _ = ops.GetSignalTarget(targetId1)
	if tgt1.Status != "new" {
		t.Fatalf("Expected target 1 to be new after unblock, got status=%s", tgt1.Status)
	}

	// Complete target 1: event should NOT be deleted yet because target 2 is still 'new'
	err = ops.TransitionSignalTargetComplete(targetId1)
	if err != nil {
		t.Fatalf("TransitionSignalTargetComplete failed: %v", err)
	}
	cleaned, err := ops.CheckAndCleanupSignalEvent(eventId)
	if err != nil {
		t.Fatalf("CheckAndCleanupSignalEvent failed: %v", err)
	}
	if cleaned {
		t.Fatalf("Event should NOT be cleaned up while target 2 is still new")
	}

	// Complete target 2: now all targets are final ('processed')
	err = ops.TransitionSignalTargetComplete(targetId2)
	if err != nil {
		t.Fatalf("TransitionSignalTargetComplete target 2 failed: %v", err)
	}
	cleaned, err = ops.CheckAndCleanupSignalEvent(eventId)
	if err != nil {
		t.Fatalf("CheckAndCleanupSignalEvent failed: %v", err)
	}
	if !cleaned {
		t.Fatalf("Event SHOULD be cleaned up when all targets are in final status")
	}

	// Verify event was deleted from SignalEvents
	evt, err := ops.GetSignalEvent(eventId)
	if err == nil && evt != nil && evt.ID != 0 {
		t.Fatalf("Expected event %d to be deleted, but still found: %v", eventId, evt)
	}
}
