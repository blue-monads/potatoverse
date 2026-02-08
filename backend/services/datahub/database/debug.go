package database

import "database/sql"

func GetDbStates(db *sql.DB) (map[string]any, error) {
	states := make(map[string]any)

	// Journal and Write-Ahead Log (WAL) related pragmas
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		return nil, err
	}
	states["journal_mode"] = journalMode

	var walAutoCheckpoint int
	if err := db.QueryRow("PRAGMA wal_autocheckpoint").Scan(&walAutoCheckpoint); err != nil {
		return nil, err
	}
	states["wal_autocheckpoint"] = walAutoCheckpoint

	var walCheckpointMode string
	if err := db.QueryRow("PRAGMA wal_checkpoint").Scan(&walCheckpointMode); err != nil {
		return nil, err
	}
	states["wal_checkpoint"] = walCheckpointMode

	// Database file and page related pragmas
	var pageSize int
	if err := db.QueryRow("PRAGMA page_size").Scan(&pageSize); err != nil {
		return nil, err
	}
	states["page_size"] = pageSize

	var cacheSize int
	if err := db.QueryRow("PRAGMA cache_size").Scan(&cacheSize); err != nil {
		return nil, err
	}
	states["cache_size"] = cacheSize

	var maxPageCount int
	if err := db.QueryRow("PRAGMA max_page_count").Scan(&maxPageCount); err != nil {
		return nil, err
	}
	states["max_page_count"] = maxPageCount

	// Performance related pragmas
	var synchronous int
	if err := db.QueryRow("PRAGMA synchronous").Scan(&synchronous); err != nil {
		return nil, err
	}
	states["synchronous"] = synchronous

	var tempStore int
	if err := db.QueryRow("PRAGMA temp_store").Scan(&tempStore); err != nil {
		return nil, err
	}
	states["temp_store"] = tempStore

	var lockingMode string
	if err := db.QueryRow("PRAGMA locking_mode").Scan(&lockingMode); err != nil {
		return nil, err
	}
	states["locking_mode"] = lockingMode

	// Foreign key and integrity related pragmas
	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		return nil, err
	}
	states["foreign_keys"] = foreignKeys == 1

	var userVersion int
	if err := db.QueryRow("PRAGMA user_version").Scan(&userVersion); err != nil {
		return nil, err
	}
	states["user_version"] = userVersion

	// Schema related pragmas
	var schemaVersion int
	if err := db.QueryRow("PRAGMA schema_version").Scan(&schemaVersion); err != nil {
		return nil, err
	}
	states["schema_version"] = schemaVersion

	var applicationId int
	if err := db.QueryRow("PRAGMA application_id").Scan(&applicationId); err != nil {
		return nil, err
	}
	states["application_id"] = applicationId

	return states, nil
}
