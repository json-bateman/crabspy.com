-- name: UpsertGameForRoom :exec
INSERT INTO games (room_id, spy_id, location, paused, timer_duration, started_at)
VALUES (?, ?, ?, 0, ?, unixepoch())
ON CONFLICT(room_id) DO UPDATE SET
    spy_id = excluded.spy_id,
    location = excluded.location,
    paused = excluded.paused,
    timer_remaining = excluded.timer_remaining,
    started_at = excluded.started_at;

-- name: SetAccusedIfAllowed :exec
UPDATE games
SET accused_id = ?
WHERE games.room_id = ?
  AND games.paused = 1
  AND games.paused_id = ?
  AND EXISTS (
    SELECT 1
    FROM room_members AS rm
    WHERE rm.room_id = games.room_id
      AND rm.user_id = ?
  );

-- name: InsertGameEvent :exec
INSERT INTO game_events (room_id, user_id, event_type, target_id, metadata, created_at)
VALUES (?, ?, ?, ?, ?, ?);

