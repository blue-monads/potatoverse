
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
  to_user text not null, 
  from_user_id text not null default 0, 
  from_project_id text not null default 0, 
  callback_token text not null default 0, 
  warn_level integer not null default 0, 
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, 
  FOREIGN KEY (to_user) REFERENCES Users(id), 
  FOREIGN KEY (from_user_id) REFERENCES Users(id), 
  FOREIGN KEY (from_project_id) REFERENCES Projects(id)
);


-- spaces

CREATE TABLE IF NOT EXISTS Spaces (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  package_id INTEGER NOT NULL,
  package_xid TEXT NOT NULL DEFAULT '',
  owns_namespace BOOLEAN NOT NULL DEFAULT FALSE,
  
  namespace_key TEXT NOT NULL DEFAULT '',
  executor_type TEXT NOT NULL DEFAULT '', 
  sub_type TEXT NOT NULL DEFAULT '',
  route_options JSON NOT NULL DEFAULT '{}',
  mcp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  mcp_definition_file TEXT NOT NULL DEFAULT '',
  mcp_options JSON NOT NULL DEFAULT '{}',
  
  overlay_for_space_id INTEGER NOT NULL DEFAULT 0,  
  owned_by INTEGER NOT NULL, 
  extrameta JSON NOT NULL DEFAULT '{}', 
  is_initilized BOOLEAN NOT NULL DEFAULT FALSE, 
  is_public BOOLEAN NOT NULL DEFAULT FALSE,

  FOREIGN KEY (owned_by) REFERENCES Users(id)
);


CREATE TABLE IF NOT EXISTS SpaceKV (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL DEFAULT '', 
  "group" TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  mod_id INTEGER NOT NULL DEFAULT 0,
  space_id INTEGER NOT NULL, -- DEFAULT 0, 
  tag1 TEXT NOT NULL DEFAULT '',
  tag2 TEXT NOT NULL DEFAULT '',
  tag3 TEXT NOT NULL DEFAULT '',
  unique(space_id, "group", key)
);


CREATE TABLE IF NOT EXISTS SpaceUsers (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  user_id INTEGER NOT NULL, 
  space_id INTEGER NOT NULL, 
  scope TEXT NOT NULL DEFAULT '', 
  extrameta JSON NOT NULL DEFAULT '{}', 
  token TEXT NOT NULL DEFAULT '', 
  FOREIGN KEY (space_id) REFERENCES Spaces(id), 
  FOREIGN KEY (user_id) REFERENCES Users(id), 
  unique(space_id, user_id)
);


CREATE TABLE IF NOT EXISTS SpaceResources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL DEFAULT '',
  space_id INTEGER NOT NULL,
  resource_type TEXT NOT NULL DEFAULT '', -- 
  resource_target TEXT NOT NULL DEFAULT '',
  attrs JSON NOT NULL DEFAULT '{}',
  unique(space_id, name)

);


-- files


CREATE TABLE IF NOT EXISTS FileShares (
  id TEXT PRIMARY KEY,
  file_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  space_id INTEGER NOT NULL default 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

