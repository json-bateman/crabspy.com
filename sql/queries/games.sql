-- name: GetCurrentGame :one
SELECT * FROM games WHERE room_id = ? ORDER BY started_at DESC LIMIT 1;

-- name: GetGameEvents :many
SELECT * FROM game_events WHERE game_id = ? ORDER BY created_at ASC;

-- name: GetGameEventsWithUsers :many
SELECT
    ge.*,
    u.username AS user_username,
    u.display_name AS user_display_name,
    u.crab_avatar AS user_crab_avatar,
    tu.username AS target_username,
    tu.display_name AS target_display_name
FROM game_events ge
LEFT JOIN users u ON u.id = ge.user_id
LEFT JOIN users tu ON tu.id = ge.target_id
WHERE ge.game_id = ?
ORDER BY ge.created_at ASC;

---------------------------
-- CREATE, UPDATE, DELETE
---------------------------
-- name: CreateGame :one
INSERT INTO games (room_id, spy_id, location, location_pool, role_assignments, timer_duration)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: InsertGameEvent :exec
INSERT INTO game_events (game_id, user_id, event_type, target_id, metadata)
VALUES (?, ?, ?, ?, ?);
