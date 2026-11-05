
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
  signal_event_key TEXT NOT NULL,
  payload BLOB NOT NULL,
  metadata JSON NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  extrameta JSON NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS SignalTargets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  signal_event_id INTEGER NOT NULL,
  signal_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'new', -- INITIAL: (new, scheduled, delayed, blocked) | FINAL: (processed, failed, expired)
  metadata JSON NOT NULL DEFAULT '{}',
  delayed_until INTEGER NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  last_retried_at INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',

  extrameta JSON NOT NULL DEFAULT '{}'
);
