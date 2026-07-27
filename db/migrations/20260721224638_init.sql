-- +goose Up
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  display TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_deleted BOOLEAN NOT NULL DEFAULT 0,
  cookie_ver INTEGER NOT NULL DEFAULT 0 -- increment on security events (like a password change or logging out all devices)
);

CREATE TABLE servers (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  owner INTEGER NOT NULL,
  FOREIGN KEY (owner) REFERENCES users(id)
);

CREATE TABLE server_members (
  server_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  is_ban BOOLEAN NOT NULL DEFAULT 0, -- track bans
  server_display TEXT, -- optional server-only display
  PRIMARY KEY (server_id, user_id),
  -- If server is deleted, wipe its memberships
  FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE server_invites (
  id string PRIMARY KEY,
  server_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  expires_at TIMESTAMP, -- can be null, no expiration
  FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE messages (
  id INTEGER PRIMARY KEY,
  server_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  content TEXT NOT NULL,
  reply_to INTEGER,
  is_reply BOOLEAN NOT NULL DEFAULT 0, -- so we can track reply to deleted message
  is_edited BOOLEAN NOT NULL DEFAULT 0,
  time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- If server is deleted, wipe its messages
  FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id),
  -- If the parent message is deleted, just set reply_to to NULL
  FOREIGN KEY (reply_to) REFERENCES messages(id) ON DELETE SET NULL
);

CREATE TABLE message_reactions (
  message_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  emoji TEXT NOT NULL,
  PRIMARY KEY (message_id, user_id, emoji),
  -- If message is deleted, delete the reactions
  FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose Down
DROP TABLE message_reactions;
DROP TABLE messages;
DROP TABLE server_members;
DROP TABLE servers;
DROP TABLE users;
