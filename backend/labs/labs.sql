



CREATE TABLE IF NOT EXISTS SpaceMcpLink (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  identity_key TEXT NOT NULL DEFAULT '',
  from_space_id INTEGER NOT NULL,
  to_space_id INTEGER NOT NULL,
  attrs JSON NOT NULL DEFAULT '{}',
  rule TEXT NOT NULL DEFAULT '',
  unique(from_space_id, to_space_id)
);

-- rMCP -> perform(action, params)
-- rMCP -> get_resource(resource_id)
-- rMCP -> list_resources(resource_type)

CREATE TABLE IF NOT EXISTS SpacePlugins (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  namespace_key TEXT NOT NULL DEFAULT '',
  executor_type TEXT NOT NULL DEFAULT '', 
  sub_type TEXT NOT NULL DEFAULT '',
  space_id INTEGER NOT NULL,
  package_id INTEGER NOT NULL,
  server_entry_file TEXT NOT NULL DEFAULT '',
  client_js_file TEXT NOT NULL DEFAULT '',
  serve_folder TEXT NOT NULL DEFAULT '', -- default is public
  trim_path_prefix TEXT NOT NULL DEFAULT '', -- default is empty
  
  unique(space_id, namespace_key)
);


CREATE TABLE IF NOT EXISTS SpaceResources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT '',
  space_id INTEGER NOT NULL,
  resource_id TEXT NOT NULL DEFAULT '',
  resource_type TEXT NOT NULL DEFAULT '', -- space, ws_room, webhook
  attrs JSON NOT NULL DEFAULT '{}',
  unique(space_id, name)
);

