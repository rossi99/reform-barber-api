-- name: CreateMedia :one
INSERT INTO media (type, entity_id, bucket_key, public_url, sort_order, alt_text, uploaded_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetMediaByID :one
SELECT * FROM media WHERE id = $1 LIMIT 1;

-- name: ListMediaByType :many
SELECT * FROM media WHERE type = $1 ORDER BY sort_order ASC, created_at ASC;

-- name: GetMediaForEntity :one
SELECT * FROM media WHERE type = $1 AND entity_id = $2 LIMIT 1;

-- name: DeleteMedia :exec
DELETE FROM media WHERE id = $1;

-- name: DeleteMediaForEntity :exec
DELETE FROM media WHERE type = $1 AND entity_id = $2;

-- name: UpdateMediaSortOrder :exec
UPDATE media SET sort_order = $2 WHERE id = $1;
