-- name: CreateBookingProduct :one
INSERT INTO booking_products (booking_id, product_id, qty, price_pence)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListBookingProducts :many
SELECT bp.*, p.name AS product_name
FROM booking_products bp
JOIN products p ON p.id = bp.product_id
WHERE bp.booking_id = $1;
