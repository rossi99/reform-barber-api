-- name: ListProducts :many
SELECT * FROM products WHERE active = true ORDER BY name ASC;

-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: CreateProduct :one
INSERT INTO products (name, price_pence)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, price_pence = $3, active = $4
WHERE id = $1
RETURNING *;
