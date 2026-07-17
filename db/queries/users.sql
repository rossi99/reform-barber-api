-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, phone, role, reminder_opt)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: UpdateUserRole :one
UPDATE users SET role = $2 WHERE id = $1 RETURNING *;

-- name: ListUsers :many
SELECT id, email, first_name, last_name, phone, role, created_at
FROM users
ORDER BY created_at DESC;

-- name: CountUsersByRole :one
SELECT COUNT(*) FROM users WHERE role = $1;
