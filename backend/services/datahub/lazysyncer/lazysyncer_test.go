package lazysyncer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/database/schema"
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

	schema := schema.Get()

	_, err = sess.SQL().Exec(schema)
	if err != nil {
		t.Fatalf("Failed to execute schema: %v", err)
	}

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

	// 4. Start LazySyncer
	// ApplyCDC logic will handle creating triggers and meta entries.
	err := ls.Start()
	if err != nil {
		t.Fatalf("Failed to start LazySyncer: %v", err)
	}

	// Verify triggers are created
	var triggerCount int
	row, err := mainDB.SQL().QueryRow("SELECT count(*) FROM sqlite_master WHERE type = 'trigger' AND name LIKE 'test_records_%'")
	if err != nil {
		t.Fatalf("Failed to query triggers: %v", err)
	}
	if err := row.Scan(&triggerCount); err != nil {
		t.Fatalf("Failed to scan trigger count: %v", err)
	}
	if triggerCount < 3 {
		t.Fatalf("Expected at least 3 triggers for test_records, got %d", triggerCount)
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

	time.Sleep(1 * time.Hour)

	close(stopCh)

	if !success {
		t.Fatalf("Failed to sync records to buddy DB")
	}
}
