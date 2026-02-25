-- name: GetCurrentGame :one
SELECT * FROM games WHERE room_id = ? ORDER BY started_at DESC LIMIT 1;

-- name: GetGameEvents :many
SELECT * FROM game_events WHERE game_id = ? ORDER BY created_at ASC;

---------------------------
-- CREATE, UPDATE, DELETE
---------------------------
-- name: CreateGame :one
INSERT INTO games (room_id, spy_id, location, timer_duration)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: InsertGameEvent :exec
INSERT INTO game_events (game_id, user_id, event_type, target_id, metadata)
VALUES (?, ?, ?, ?, ?);
