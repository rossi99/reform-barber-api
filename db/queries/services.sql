-- name: ListServices :many
SELECT * FROM services WHERE active = true ORDER BY num ASC;

-- name: ListAllServices :many
SELECT * FROM services ORDER BY num ASC;

-- name: GetServiceByID :one
SELECT * FROM services WHERE id = $1 LIMIT 1;

-- name: CreateService :one
INSERT INTO services (num, name, name_html, description, duration_mins, price_pence)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateService :one
UPDATE services
SET num = $2, name = $3, name_html = $4, description = $5,
    duration_mins = $6, price_pence = $7, active = $8
WHERE id = $1
RETURNING *;

-- name: SetServicePublished :one
UPDATE services SET active = $2 WHERE id = $1 RETURNING *;
