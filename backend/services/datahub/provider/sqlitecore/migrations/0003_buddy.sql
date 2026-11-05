
CREATE TABLE IF NOT EXISTS SelfCDCMeta (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  table_name TEXT NOT NULL,
  primary_key TEXT NOT NULL,
  
  current_max_cdc_id INTEGER NOT NULL DEFAULT 0,

  gc_max_records INTEGER NOT NULL DEFAULT 0,
  last_gc_at INTEGER NOT NULL DEFAULT 0,
  last_current_cached_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  current_schema_hash TEXT NOT NULL DEFAULT '',

  extrameta JSON NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS BuddyCDCMeta (
  id INTEGER PRIMARY KEY,
  pubkey TEXT NOT NULL,
  remote_table_id INTEGER NOT NULL,
  table_name TEXT NOT NULL,

  current_max_cdc_id INTEGER NOT NULL DEFAULT 0,
  synced_cdc_id INTEGER NOT NULL DEFAULT 0,

  current_schema_hash TEXT NOT NULL DEFAULT '',
  is_deleted BOOLEAN NOT NULL DEFAULT 0,
  extrameta JSON NOT NULL DEFAULT '{}'
);
