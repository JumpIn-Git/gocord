/* USERS */
-- name: CreateUser :exec
INSERT INTO users(id, username, display, password_hash) VALUES (?, ?, ?, ?);

-- name: DeleteUser :exec
UPDATE users
SET
  username = 'Deleted User ' || id,
  display = 'Deleted User ' || id,
  password_hash = '',
  is_deleted = 1
WHERE
  id = ?
  AND is_deleted = 0;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? AND is_deleted = 0;

-- name: GetUserServersIDs :many
SELECT server_id FROM server_members WHERE user_id = ? AND is_ban = 0;

-- name: UserInServer :one
SELECT
  EXISTS (
    SELECT 1
    FROM server_members
    WHERE
      user_id = ?
      AND server_id = ?
      AND is_ban = 0
  );

-- name: IncCookieVer :exec
UPDATE users SET cookie_ver = cookie_ver + 1 WHERE id = ?;

/* SERVERS */
-- name: CreateServer :exec
INSERT INTO servers(id, name, owner) VALUES (?, ?, ?);

-- name: DeleteServer :exec
DELETE FROM servers WHERE id = ?;

-- name: JoinServer :exec
INSERT INTO server_members(user_id, server_id) VALUES (?, ?);

-- name: LeaveServer :exec
DELETE FROM server_members WHERE user_id = ? AND server_id = ? AND is_ban = 0;

-- name: GetServerMembers :many
SELECT * FROM server_members WHERE server_id = ? LIMIT ? OFFSET ?;

-- name: GetServerMemberCount :one
SELECT COUNT(*) FROM server_members WHERE server_id = ?;

-- name: CreateMembership :exec
INSERT INTO server_members(user_id, server_id) VALUES (?, ?);

-- name: IsOwner :one
SELECT EXISTS (SELECT 1 FROM servers WHERE id = ? AND owner = ?);

/* MESSAGES */
-- name: CreateMessage :exec
INSERT INTO messages(id, server_id, user_id, content, reply_to, is_reply)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetServerMessages :many
SELECT * FROM messages WHERE server_id = ? ORDER BY id LIMIT ? OFFSET ?;

-- name: GetServerMessage :one
SELECT * FROM messages WHERE server_id = ? AND id = ?;

-- name: EditMessage :execrows
UPDATE messages SET content = ?, is_edited = 1 WHERE id = ? AND user_id = ?;

-- name: DeleteMessage :execrows
DELETE FROM messages
WHERE
  messages.id = ?1
  AND (user_id = ?2 OR server_id IN (SELECT id FROM servers WHERE owner = ?2)); -- Server owner can delete any message

/* REACTIONS */
-- name: CreateReaction :exec
INSERT INTO message_reactions(message_id, user_id, emoji) VALUES (?, ?, ?);

-- name: GetMessageReactions :many
SELECT * FROM message_reactions WHERE message_id = ?;

-- name: DeleteReaction :execrows
DELETE FROM message_reactions WHERE message_id = ? AND user_id = ? AND emoji = ?;

/* INVITES */
-- name: DeleteExpiredInvites :exec
-- Run this periodically to delete expired invites
DELETE FROM server_invites WHERE
  expires_at IS NOT NULL
  AND expires_at <= datetime('now');

-- name: GetInvite :one
SELECT * FROM server_invites WHERE id = ?;

-- name: DeleteInvite :exec
DELETE FROM server_invites WHERE id = ?;

-- name: MakeInvite :exec
INSERT INTO server_invites(id, server_id, user_id, expires_at)
VALUES (?, ?, ?, ?);
