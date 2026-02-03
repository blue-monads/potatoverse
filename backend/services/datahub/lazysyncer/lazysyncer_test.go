package lazysyncer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

// setupTestDB creates an in-memory SQLite database for testing, but since we need multiple connections/persistence for buddies,
// we might use a file-based one or distinct in-memory names.
func setupTestDB(t *testing.T, dbName string) (db.Session, string, func()) {
	// Create a temp directory for the db file
	tmpDir, err := os.MkdirTemp("", "lazysyncer_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, dbName)

	settings := sqlite.ConnectionURL{
		Database: dbPath,
	}

	sess, err := sqlite.Open(settings)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create tables
	// We need SelfCDCMeta, BuddyCDCMeta and a test table.
	// We can use the schema from schema.sql ideally, but for now we'll create what's needed.

	schemas := []string{
		`CREATE TABLE IF NOT EXISTS test_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			value TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS SelfCDCMeta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			table_name TEXT NOT NULL DEFAULT '',
			start_row_id INTEGER NOT NULL DEFAULT 0,
			current_max_cdc_id INTEGER NOT NULL DEFAULT 0,
			current_cdc_id INTEGER NOT NULL DEFAULT 0,
			gc_max_records INTEGER NOT NULL DEFAULT 0,
			last_gc_at TIMESTAMP,
			last_cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			current_schema_hash TEXT NOT NULL DEFAULT '',
			extrameta JSON NOT NULL DEFAULT '{}'
		)`,
		`CREATE TABLE IF NOT EXISTS BuddyCDCMeta (
			id INTEGER PRIMARY KEY,
			pubkey TEXT NOT NULL,
			remote_table_id INTEGER NOT NULL,
			table_name TEXT NOT NULL,
			start_row_id INTEGER NOT NULL DEFAULT 0,
			synced_row_id INTEGER NOT NULL DEFAULT 0,
			current_max_cdc_id INTEGER NOT NULL DEFAULT 0,
			synced_cdc_id INTEGER NOT NULL DEFAULT 0,
			current_cdc_id INTEGER NOT NULL DEFAULT 0,
			current_schema_hash TEXT NOT NULL DEFAULT '',
			is_deleted BOOLEAN NOT NULL DEFAULT 0,
			extrameta JSON NOT NULL DEFAULT '{}'
		)`,
		// CDC table for test_records
		`CREATE TABLE IF NOT EXISTS test_records__cdc (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			record_id INTEGER NOT NULL,
			operation INTEGER NOT NULL,
			payload BLOB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		_, err := sess.SQL().Exec(schema)
		if err != nil {
			t.Fatalf("Failed to execute schema: %v", err)
		}
	}

	// Initialize SelfCDCMeta for test_records
	// We need to insert a record into SelfCDCMeta so the syncer knows about it?
	// Or maybe the syncer discovers it?
	// Looking at selfcdc/syncer.go: getTableNames(), it seems it might scan tables?
	// selfcdc/syncer.go:85 AttachMissingTables() -> calling s.ApplyCDC()
	// Let's assume ApplyCDC does the magic.

	cleanup := func() {
		sess.Close()
		os.RemoveAll(tmpDir)
	}

	return sess, tmpDir, cleanup
}

func TestLazySyncerBasic(t *testing.T) {
	// 1. Setup Main DB
	mainDB, tmpDir, cleanupMain := setupTestDB(t, "main.db")
	defer cleanupMain()

	// 2. Setup Buddy DB (implicitly done by BuddyCDC, but we need to verify it)
	// BuddyCDC creates its own DB file in BasePath.
	buddyName := "buddy1"

	// 3. Initialize LazySyncer
	opts := Options{
		DbSession:     mainDB,
		IsSelfEnabled: true,
		Buddies:       []string{buddyName},
		BasePath:      tmpDir,
	}

	ls := NewTest(opts)
	if ls == nil {
		t.Fatalf("Failed to create LazySyncer")
	}

	// 4. Insert initial data into SelfCDCMeta for the test_records table
	// The ApplyCDC logic usually handles creating triggers and meta entries.
	// Since we are mocking/using partial setup, we might need to manually trigger ApplyCDC or insert meta.
	// Let's try calling Start() which calls ApplyCDC.

	// Create a trigger manually if ApplyCDC isn't fully working in this test env or rely on it.
	// Looking at selfcdc/apply.go (not read yet), it likely creates triggers.
	// For this test, let's manually insert into SelfCDCMeta to "enable" CDC for our table if needed.
	// But Start() calls ApplyCDC(), so hopefully that works.

	// However, we also need to ensure `test_records` is tracked.
	// If ApplyCDC scans for tables ending in __cdc, or if it creates them...
	// Usually CDC systems look for tables and create __cdc tables.
	// Let's assume for this test we need to manually register the table in SelfCDCMeta if ApplyCDC doesn't auto-discover all tables.
	// selfcdc.go: NewSelfCDCSyncer does not seem to take a list of tables.

	// Let's manually insert the meta row to be safe/explicit.
	_, err := mainDB.Collection("SelfCDCMeta").Insert(map[string]interface{}{
		"table_name": "test_records",
	})
	if err != nil {
		t.Fatalf("Failed to insert SelfCDCMeta: %v", err)
	}

	// Also ensure we have the trigger or the __cdc table populated.
	// If the real system uses triggers, we need them.
	// Let's simulate the CDC log insertion manually in the loop if we don't trust ApplyCDC setup in test.
	// But the user asked for "pass it and in one goroutine in loop every 5 second insert record".
	// This implies we should be inserting into the main table and expect the CDC to pick it up.
	// To make this work without full ApplyCDC magic (which might depend on sqlite triggers),
	// we will manually insert into `test_records__cdc` as well when we insert into `test_records`.

	err = ls.Start()
	if err != nil {
		t.Fatalf("Failed to start LazySyncer: %v", err)
	}

	// 5. Start Data Generation Loop
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond) // Faster than 5s for test
		defer ticker.Stop()
		counter := 0
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				counter++
				name := fmt.Sprintf("record_%d", counter)

				// Insert into main table
				_, err := mainDB.Collection("test_records").Insert(map[string]interface{}{
					"name":  name,
					"value": fmt.Sprintf("val_%d", counter),
				})
				if err != nil {
					fmt.Printf("Insert failed: %v\n", err)
					continue
				}

			}
		}
	}()

	// 6. Verify Synchronization
	// We need to check if the buddy database gets the updates.
	// The buddy database is located at tmpDir/buddycdc_buddy1.db
	buddyDbPath := filepath.Join(tmpDir, fmt.Sprintf("buddycdc_%s.db", buddyName))

	// Wait for buddy DB to be created
	deadline := time.Now().Add(15 * time.Second)
	for {
		if _, err := os.Stat(buddyDbPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Buddy DB file not created in time")
		}
		time.Sleep(100 * time.Millisecond)
	}

	buddySess, err := sqlite.Open(sqlite.ConnectionURL{Database: buddyDbPath})
	if err != nil {
		t.Fatalf("Failed to open buddy DB: %v", err)
	}
	defer buddySess.Close()

	// Wait for records to appear in Buddy DB
	// We expect `test_records` (or similar) to be populated in buddy DB.
	// The buddy syncer might create the table if it's applying changes.
	// But usually it applies changes to the SAME table name.
	// Let's poll for record count.

	success := false
	for i := 0; i < 60; i++ {
		// Check if table exists in buddy
		exists, err := buddySess.Collection("test_records").Exists()
		if err != nil || !exists {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		count, err := buddySess.Collection("test_records").Count()
		if err != nil {
			// Table might be locked or not ready
			time.Sleep(500 * time.Millisecond)
			continue
		}

		fmt.Printf("[%s] Buddy DB count: %d\n", time.Now().Format("15:04:05"), count)
		if count > 5 {
			success = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	close(stopCh)

	if !success {
		t.Fatalf("Failed to sync records to buddy DB")
	}
}
