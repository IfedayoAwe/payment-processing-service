-- name: GetUserByID :one
SELECT * FROM users
WHERE user_id = $1;

-- name: CreateUser :one
INSERT INTO users (user_id)
VALUES ($1)
RETURNING *;
