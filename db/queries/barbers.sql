-- name: ListBarbers :many
SELECT b.*, u.email FROM barbers b
JOIN users u ON u.id = b.user_id
WHERE b.active = true
ORDER BY b.num ASC;

-- name: GetBarberByID :one
SELECT b.*, u.email FROM barbers b
JOIN users u ON u.id = b.user_id
WHERE b.id = $1;

-- name: GetBarberByUserID :one
SELECT * FROM barbers WHERE user_id = $1 LIMIT 1;

-- name: CreateBarber :one
INSERT INTO barbers (user_id, name, title, bio, num)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateBarber :one
UPDATE barbers
SET name = $2, title = $3, bio = $4, num = $5, active = $6
WHERE id = $1
RETURNING *;

-- name: ListActiveBarberIDs :many
SELECT id FROM barbers WHERE active = true ORDER BY num ASC;
