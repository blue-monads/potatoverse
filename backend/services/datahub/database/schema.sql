
CREATE TABLE IF NOT EXISTS GlobalConfig (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  key TEXT NOT NULL DEFAULT '', 
  group_name TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  unique(group_name, key)
);


CREATE TABLE IF NOT EXISTS Users (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  username TEXT,
  email TEXT, 
  phone TEXT,

  name TEXT NOT NULL, 
  utype TEXT NOT NULL DEFAULT 'real',  -- admin, normal, bot
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
  invited_as_type TEXT NOT NULL DEFAULT 'normal', -- user, admin, moderator, developer
  expires_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  unique(email)
);

CREATE TABLE IF NOT EXISTS UserConfig (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  key TEXT NOT NULL DEFAULT '', 
  group_name TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  user_id INTEGER NOT NULL, 
  unique(user_id, group_name, key),
  FOREIGN KEY (user_id) REFERENCES Users(id)
);



CREATE TABLE IF NOT EXISTS UserDevices (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  name TEXT NOT NULL DEFAULT '', 
  dtype TEXT NOT NULL DEFAULT 'sesssion', --  session token
  token_hash TEXT NOT NULL DEFAULT '', 
  user_id INTEGER NOT NULL, 
  project_id INTEGER NOT NULL DEFAULT 0,
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
  name TEXT NOT NULL DEFAULT '', 
  info TEXT NOT NULL DEFAULT '', 
  stype TEXT NOT NULL DEFAULT '', 
  owned_by INTEGER NOT NULL, 
  extrameta JSON NOT NULL DEFAULT '{}', 
  is_initilized BOOLEAN NOT NULL DEFAULT FALSE, 
  is_public BOOLEAN NOT NULL DEFAULT FALSE,
  FOREIGN KEY (owned_by) REFERENCES Users(id)
);

CREATE TABLE IF NOT EXISTS SpaceConfig (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  -- bprint_id TEXT NOT NULL DEFAULT '',
  key TEXT NOT NULL DEFAULT '', 
  group_name TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  space_id INTEGER NOT NULL, -- DEFAULT 0, 
  unique(space_id, group_name, key),
  FOREIGN KEY (space_id) REFERENCES Spaces(id)
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

-- files

CREATE TABLE IF NOT EXISTS Files (
  id INTEGER PRIMARY KEY AUTOINCREMENT, 
  name TEXT NOT NULL DEFAULT '', 
  is_folder BOOLEAN NOT NULL DEFAULT FALSE,
  path TEXT NOT NULL DEFAULT '', 
  size INTEGER NOT NULL DEFAULT 0, 
  mime TEXT NOT NULL DEFAULT '', 
  hash TEXT NOT NULL DEFAULT '',
  storeType INTEGER NOT NULL DEFAULT 0, -- 0: inline_blob, 1: external_blob, 2: mulit_part_blob
  preview BLOB, 
  blob BLOB,
  external BOOLEAN NOT NULL DEFAULT FALSE,
  owner_space_id INTEGER NOT NULL DEFAULT 0,
  created_by INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (owner_space_id, path, name)
);



CREATE TABLE IF NOT EXISTS FilePartedBlobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_id INTEGER NOT NULL,
  size INTEGER NOT NULL,
  part_id INTEGER NOT NULL,
  blob BLOB NOT NULL
);



CREATE TABLE IF NOT EXISTS FileShares (
  id TEXT PRIMARY KEY,
  file_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  space_id INTEGER NOT NULL default 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS BprintInstalls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slug TEXT NOT NULL, 
  type TEXT NOT NULL DEFAULT 'db', -- db, file-open, file-zip etc.
  reference TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL DEFAULT '',
  info TEXT NOT NULL DEFAULT '',
  tags TEXT NOT NULL DEFAULT '',
  
  installed_by INTEGER NOT NULL DEFAULT 0,
  installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (slug)
);


CREATE TABLE IF NOT EXISTS BprintInstalls (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  bprint_slug TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '', 
  is_folder BOOLEAN NOT NULL DEFAULT FALSE,
  path TEXT NOT NULL DEFAULT '', 
  size INTEGER NOT NULL DEFAULT 0, 
  mime TEXT NOT NULL DEFAULT '', 
  hash TEXT NOT NULL DEFAULT '',
  storeType INTEGER NOT NULL DEFAULT 0, -- 0: inline_blob, 1: external_blob, 2: mulit_part_blob
  preview BLOB, 
  blob BLOB,
  external BOOLEAN NOT NULL DEFAULT FALSE,
  created_by INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (bprint_slug, path, name)
);


CREATE TABLE IF NOT EXISTS BprintInstallFileBlobs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_id INTEGER NOT NULL,
  size INTEGER NOT NULL,
  part_id INTEGER NOT NULL,
  blob BLOB NOT NULL
);