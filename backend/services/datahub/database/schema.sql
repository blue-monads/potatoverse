
CREATE TABLE IF NOT EXISTS GlobalConfig (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  key TEXT NOT NULL DEFAULT '', 
  "group" TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  unique("group", key)
);

CREATE TABLE IF NOT EXISTS UserGroups (
  name TEXT PRIMARY KEY,  
  info TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  extrameta JSON NOT NULL DEFAULT '{}',
  UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS Users (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  username TEXT,
  email TEXT, 
  phone TEXT,

  name TEXT NOT NULL, 
  utype TEXT NOT NULL DEFAULT 'user', -- user, bot, api
  ugroup TEXT NOT NULL, --  UserGroups.name
  bio TEXT NOT NULL DEFAULT '', 
  password TEXT NOT NULL, 
  is_verified BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  owner_user_id INTEGER NOT NULL DEFAULT 0,
  owner_space_id INTEGER NOT NULL DEFAULT 0,
  extrameta JSON NOT NULL DEFAULT '{}',
  msg_read_head INTEGER NOT NULL DEFAULT 0,
  
  disabled BOOLEAN NOT NULL DEFAULT FALSE, 
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  unique(username),  
  unique(email),
  unique(phone)
);


CREATE TABLE IF NOT EXISTS UserInvites (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  email TEXT NOT NULL DEFAULT '', 
  role TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending', -- pending, accepted, rejected
  invited_by INTEGER NOT NULL DEFAULT 0,
  invited_as_type TEXT NOT NULL DEFAULT 'user', -- user, admin, moderator, developer
  expires_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  unique(email)
);

CREATE TABLE IF NOT EXISTS UserConfig (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  key TEXT NOT NULL DEFAULT '', 
  "group" TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  user_id INTEGER NOT NULL, 
  unique(user_id, "group", key),
  FOREIGN KEY (user_id) REFERENCES Users(id)
);



CREATE TABLE IF NOT EXISTS UserDevices (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  name TEXT NOT NULL DEFAULT '', 
  dtype TEXT NOT NULL DEFAULT 'sesssion', --  session token
  token_hash TEXT NOT NULL DEFAULT '', 
  user_id INTEGER NOT NULL, 
  last_ip TEXT NOT NULL DEFAULT '',
  last_login TEXT NOT NULL DEFAULT '',
  extrameta JSON NOT NULL DEFAULT '{}', 
  expires_on TIMESTAMP not null, 
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES Users(id)
);

CREATE TABLE IF NOT EXISTS UserMessages(
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  title text not null default '', 
  is_read boolean not null default FALSE, 
  type text not null default "messsage", 
  contents text not null, 
  to_user INTEGER not null default 0, 
  from_user_id INTEGER not null default 0, 
  from_space_id INTEGER not null default 0, 
  callback_token TEXT not null default '', 
  warn_level integer not null default 0, 
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- spaces

CREATE TABLE IF NOT EXISTS PackageInstalls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT '',  
  install_repo TEXT NOT NULL DEFAULT '',
  canonical_url TEXT NOT NULL DEFAULT '',
  storage_type TEXT NOT NULL DEFAULT 'db', -- db, file-open, file-zip etc.
  active_install_id INTEGER NOT NULL DEFAULT 0,
  installed_by INTEGER NOT NULL DEFAULT 0,
  installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_active BOOLEAN NOT NULL DEFAULT FALSE,
  dev_token TEXT NOT NULL DEFAULT ''
);


CREATE TABLE IF NOT EXISTS PackageVersion (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  install_id INTEGER NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  slug TEXT NOT NULL DEFAULT '',
  info TEXT NOT NULL DEFAULT '',
  tags TEXT NOT NULL DEFAULT '',
  spec_file TEXT NOT NULL DEFAULT '',
  format_version TEXT NOT NULL DEFAULT '',
  author_name TEXT NOT NULL DEFAULT '',
  author_email TEXT NOT NULL DEFAULT '',
  author_site TEXT NOT NULL DEFAULT '',
  source_code TEXT NOT NULL DEFAULT '',
  license TEXT NOT NULL DEFAULT '',
  version TEXT NOT NULL DEFAULT '',
  init_page TEXT NOT NULL DEFAULT '',
  update_page TEXT NOT NULL DEFAULT ''
);



CREATE TABLE IF NOT EXISTS Spaces (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  install_id INTEGER NOT NULL,  
  namespace_key TEXT NOT NULL DEFAULT '',
  space_type TEXT NOT NULL DEFAULT '', -- App, AppOverlay, AppPlugin
  executor_type TEXT NOT NULL DEFAULT '', 
  executor_sub_type TEXT NOT NULL DEFAULT '',
  route_options JSON NOT NULL DEFAULT '{}',
  server_file TEXT NOT NULL DEFAULT '',
  
  overlay_for_space_id INTEGER NOT NULL DEFAULT 0,  
  owned_by INTEGER NOT NULL, 
  extrameta JSON NOT NULL DEFAULT '{}', 
  is_initilized BOOLEAN NOT NULL DEFAULT FALSE, 
  is_public BOOLEAN NOT NULL DEFAULT FALSE
);


CREATE TABLE IF NOT EXISTS SpaceKV (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL DEFAULT '', 
  "group" TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  mod_id INTEGER NOT NULL DEFAULT 0,
  install_id INTEGER NOT NULL, -- DEFAULT 0, 
  tag1 TEXT NOT NULL DEFAULT '',
  tag2 TEXT NOT NULL DEFAULT '',
  tag3 TEXT NOT NULL DEFAULT '',
  unique(install_id, "group", key)
);


CREATE TABLE IF NOT EXISTS SpaceUsers (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  user_id INTEGER NOT NULL,
  install_id INTEGER NOT NULL,
  space_id INTEGER NOT NULL DEFAULT 0, 
  scope TEXT NOT NULL DEFAULT '', 
  extrameta JSON NOT NULL DEFAULT '{}', 
  token TEXT NOT NULL DEFAULT '',
  unique(install_id, space_id, user_id)
);


CREATE TABLE IF NOT EXISTS SpaceCapabilities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT '',
  install_id INTEGER NOT NULL,
  space_id INTEGER NOT NULL DEFAULT 0,
  capability_type TEXT NOT NULL DEFAULT '',
  options JSON NOT NULL DEFAULT '{}',
  extrameta JSON NOT NULL DEFAULT '{}',
  unique(install_id, space_id, name)

);


CREATE TABLE IF NOT EXISTS MQSubscriptions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  install_id INTEGER NOT NULL,
  space_id INTEGER NOT NULL DEFAULT 0,
  event_key TEXT NOT NULL DEFAULT '',
  target_type TEXT NOT NULL DEFAULT '', -- webhook, script, space_method
  target_space_id INTEGER NOT NULL DEFAULT 0,
  target_endpoint TEXT NOT NULL DEFAULT '',
  target_options JSON NOT NULL DEFAULT '{}', -- it has creds, api keys and other options
  target_code TEXT NOT NULL DEFAULT '',
  rules JSON NOT NULL DEFAULT '{}',
  transform JSON NOT NULL DEFAULT '{}',
  delay_start INTEGER NOT NULL DEFAULT 0,
  retry_delay INTEGER NOT NULL DEFAULT 0,
  max_retries INTEGER NOT NULL DEFAULT 0,
  expires_on INTEGER NOT NULL DEFAULT 0, -- (created_at + expires_in > now) then status is expired
  collapse_interval INTEGER NOT NULL DEFAULT 0, -- 1 minute, 5 minute, 15 minute etc in seconds
  extrameta JSON NOT NULL DEFAULT '{}',
  created_by INTEGER NOT NULL DEFAULT 0,
  disabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- utc timestamp rounded to nearsest interval

CREATE TABLE IF NOT EXISTS MQEvents (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  install_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  payload BLOB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status TEXT NOT NULL DEFAULT 'new', -- new, scheduled, processed
  extrameta JSON NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS MQEventTargets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  collapse_key TEXT NOT NULL DEFAULT '',
  event_id INTEGER NOT NULL,
  subscription_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'new', -- new, processing, start_delayed, delayed, processed, failed, expired
  delayed_until INTEGER NOT NULL DEFAULT 0,
  retry_count INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',
  extrameta JSON NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS CDCMeta (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  table_name TEXT NOT NULL DEFAULT '',
  -- cdc_start_id rowid of the first record in the table, all records before this id has to be synced before syncing from cdc
  cdc_start_id INTEGER NOT NULL DEFAULT 0, 
  current_cdc_id INTEGER NOT NULL DEFAULT 0,
  gc_max_records INTEGER NOT NULL DEFAULT 0,
  last_gc_at INTEGER NOT NULL DEFAULT 0,
  extrameta JSON NOT NULL DEFAULT '{}'
)