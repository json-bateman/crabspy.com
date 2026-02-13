-- name: GetAllRooms :many
SELECT * FROM rooms;

-- name: GetRoomsAndMembers :many
SELECT rooms.*, COUNT(rm.user_id) AS player_count
FROM rooms
LEFT JOIN room_members AS rm ON rm.room_id = rooms.id
GROUP BY rooms.id;

-- name: GetRoomById :one
SELECT * FROM rooms WHERE id = ?;

-- name: GetRoomByCode :one
SELECT * FROM rooms WHERE code = ?;

-- name: GetRoomMembers :many
SELECT users.id, users.username, users.display_name, rm.is_ready
FROM room_members rm
JOIN users ON users.id = rm.user_id
WHERE rm.room_id = ?;

-- name: GetGameByRoomID :one
SELECT * FROM games WHERE room_id = ?;

---------------------------
-- CREATE, UPDATE, DELETE
---------------------------
-- name: CreateRoom :one
INSERT INTO rooms (name, host_id, max_players, max_locations, code)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: JoinRoom :exec
INSERT OR IGNORE INTO room_members (room_id, user_id) VALUES (?, ?);

-- name: LeaveRoom :exec
DELETE FROM room_members WHERE room_id = ? AND user_id = ?;

-- name: UpdateRoomHost :exec
UPDATE rooms SET host_id = ? WHERE id = ?;

-- name: UpdateRoomState :exec
UPDATE rooms SET state = ? WHERE id = ?;

-- name: UpsertGameForRoom :exec
INSERT INTO games (room_id, spy_id, location, paused, started_at)
VALUES (?, ?, ?, 0, unixepoch())
ON CONFLICT(room_id) DO UPDATE SET
    spy_id = excluded.spy_id,
    location = excluded.location,
    paused = excluded.paused,
    started_at = excluded.started_at;

