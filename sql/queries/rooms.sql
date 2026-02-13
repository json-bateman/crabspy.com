---------
-- READ
---------
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

