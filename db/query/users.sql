-- name: GetUserByID :one
SELECT user_id, name, type, pin_hash, created_at, updated_at
FROM users
WHERE user_id = $1;
