-- name: UpsertGameForRoom :exec
INSERT INTO games (room_id, spy_id, location, paused, timer_remaining, started_at)
VALUES (?, ?, ?, 0, ?, unixepoch())
ON CONFLICT(room_id) DO UPDATE SET
    spy_id = excluded.spy_id,
    location = excluded.location,
    paused = excluded.paused,
    timer_remaining = excluded.timer_remaining,
    started_at = excluded.started_at;

-- name: UpdateGameTimer :exec
UPDATE games
SET timer_remaining = ?
WHERE room_id = ?;

-- name: TogglePauseGame :exec
UPDATE games 
SET 
    paused = 1 - paused,
    paused_id = ?,
    accused_id = NULL

WHERE room_id = ?
RETURNING *;

-- name: UpdateAccused :exec
UPDATE games
SET accused_id = ?
WHERE room_id = ?;

