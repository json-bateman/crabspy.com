-- name: GetAllRooms :many
SELECT * FROM rooms;

-- name: CreateRoom :one
INSERT INTO rooms (name, host_id, max_players, max_locations)
VALUES (?, ?, ?, ?)
RETURNING *;


 
