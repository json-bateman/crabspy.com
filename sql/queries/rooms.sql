-- name: GetAllRooms :many
SELECT * FROM rooms;

-- name: GetRoomsAndMembers :many
SELECT rooms.*, COUNT(rm.user_id) AS player_count
FROM rooms
LEFT JOIN room_members AS rm ON rm.room_id = rooms.id
GROUP BY rooms.id;

-- name: GetRoomById :one
SELECT * FROM rooms WHERE id = ?;

-- name: CreateRoom :one
INSERT INTO rooms (name, host_id, max_players, max_locations)
VALUES (?, ?, ?, ?)
RETURNING *;


 
