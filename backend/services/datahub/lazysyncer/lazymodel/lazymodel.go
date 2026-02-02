package lazymodel

const CDCTableTemplate = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL, -- 0: insert, 1: update, 2: delete, 3: schema_init 4:schema_change
	schema_text TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
