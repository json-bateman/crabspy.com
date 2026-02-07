-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ?
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, password, display_name)
VALUES (?, ?, ?)
RETURNING *;
 
