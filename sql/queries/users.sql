-- name: GetUserById :one
SELECT * FROM users
WHERE id = ?
LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ?
LIMIT 1;

---------------------------
-- CREATE, UPDATE, DELETE
---------------------------
-- name: CreateUser :one
INSERT INTO users (username, password_hash, display_name)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateUserAvatar :one
UPDATE users SET crab_avatar = ? WHERE id = ? RETURNING *;
