

CREATE TABLE IF NOT EXISTS MQEvents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  install_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  payload BLOB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  status TEXT NOT NULL DEFAULT 'pending', -- new, scheduled, processed
  targets JSON NOT NULL DEFAULT '[]',
  extrameta JSON NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS MQEventTargets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_id INTEGER NOT NULL,
  target_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'new', -- new, processing, delayed, processed
  delayed_until INTEGER NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',
  extrameta JSON NOT NULL DEFAULT '{}'
)