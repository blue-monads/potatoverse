

CREATE TABLE IF NOT EXISTS MQEvents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  install_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  payload BLOB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status INTEGER NOT NULL DEFAULT 0, -- 0:pending, 1:processing, 2:processed, 3:delayed, 4:failed
  delay_until INTEGER NOT NULL DEFAULT 0,
  delayed TEXT NOT NULL DEFAULT '{}',
  processed TEXT NOT NULL DEFAULT '{}',
  errors TEXT NOT NULL DEFAULT '{}'
);