package buddycdc

const TemplateTable = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL, -- 0: insert, 1: update, 2: delete
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
const MetaTable = `
CREATE TABLE IF NOT EXISTS BuddyCDCMeta (
  id INTEGER PRIMARY KEY,
  remote_table_id INTEGER NOT NULL,
  table_name TEXT NOT NULL,
  cdc_start_id INTEGER NOT NULL DEFAULT 0,
  current_cdc_id INTEGER NOT NULL DEFAULT 0,
  extrameta JSON NOT NULL DEFAULT '{}'
);

`
