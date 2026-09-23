
CREATE TABLE IF NOT EXISTS Signals (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  signal_key TEXT NOT NULL,

  emitter_install_id INTEGER NOT NULL,
  emitter_space_id INTEGER NOT NULL,

  receiver_install_id INTEGER NOT NULL,
  receiver_space_id INTEGER NOT NULL,
  receiver_handler TEXT NOT NULL,

  managed_by TEXT NOT NULL DEFAULT 'both', -- emitter, receiver, both

  expires_on INTEGER NOT NULL DEFAULT 0,
  max_retries INTEGER NOT NULL DEFAULT 0,
  retry_delay INTEGER NOT NULL DEFAULT 0,

  created_by INTEGER NOT NULL DEFAULT 0,
  disabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS SignalEvents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  signal_id INTEGER NOT NULL,
  payload BLOB NOT NULL,
  metadata JSON NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status TEXT NOT NULL DEFAULT 'new', -- new, scheduled, processed, failed, delayed, expired
  delayed_until INTEGER NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  last_retried_at INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',

  extrameta JSON NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS SigalEvents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  signal_id INTEGER NOT NULL,
  payload BLOB NOT NULL,
  metadata JSON NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status TEXT NOT NULL DEFAULT 'new', -- new, scheduled, processed, failed, delayed, expired
  delayed_until INTEGER NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  last_retried_at INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',

  extrameta JSON NOT NULL DEFAULT '{}'
);
