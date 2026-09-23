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

	// Test QuerySignalEvents
	events, err := ops.QuerySignalEvents(1, 0, "", 10, 0)
	if err != nil {
		t.Fatalf("QuerySignalEvents failed: %v", err)
	}
	t.Logf("Events: %v", events)

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
}
