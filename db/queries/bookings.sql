-- name: CreateBooking :one
INSERT INTO bookings (reference, user_id, barber_id, service_id, date, time_start, time_end, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'confirmed')
RETURNING *;

-- name: GetBookingByID :one
SELECT * FROM bookings WHERE id = $1 LIMIT 1;

-- name: GetBookingByReference :one
SELECT * FROM bookings WHERE reference = $1 LIMIT 1;

-- name: ListBookingsForUser :many
SELECT b.*, bar.name AS barber_name, s.name AS service_name
FROM bookings b
JOIN barbers bar ON bar.id = b.barber_id
JOIN services s ON s.id = b.service_id
WHERE b.user_id = $1
ORDER BY b.date DESC, b.time_start DESC;

-- name: ListBookingsForBarber :many
SELECT b.*, u.first_name, u.last_name, u.phone, s.name AS service_name, s.duration_mins
FROM bookings b
LEFT JOIN users u ON u.id = b.user_id
JOIN services s ON s.id = b.service_id
WHERE b.barber_id = $1 AND b.date = $2 AND b.status != 'cancelled'
ORDER BY b.time_start ASC;

-- name: ListBookingsForBarberDateRange :many
SELECT b.*, u.first_name, u.last_name, s.name AS service_name, s.duration_mins
FROM bookings b
LEFT JOIN users u ON u.id = b.user_id
JOIN services s ON s.id = b.service_id
WHERE b.barber_id = $1 AND b.date BETWEEN $2 AND $3 AND b.status != 'cancelled'
ORDER BY b.date ASC, b.time_start ASC;

-- Confirmed bookings for a barber on a date - used by availability engine.
-- name: ListConfirmedBookingsForBarberDate :many
SELECT time_start, time_end FROM bookings
WHERE barber_id = $1 AND date = $2 AND status = 'confirmed';

-- name: CancelBooking :one
UPDATE bookings
SET status = 'cancelled', cancelled_at = now()
WHERE id = $1 AND status = 'confirmed'
RETURNING *;

-- name: ListAllBookings :many
SELECT b.*, bar.name AS barber_name, u.first_name, u.last_name, s.name AS service_name
FROM bookings b
JOIN barbers bar ON bar.id = b.barber_id
LEFT JOIN users u ON u.id = b.user_id
JOIN services s ON s.id = b.service_id
WHERE (sqlc.narg(status)::text IS NULL OR b.status = sqlc.narg(status))
  AND (sqlc.narg(barber_id)::uuid IS NULL OR b.barber_id = sqlc.narg(barber_id))
  AND (sqlc.narg(date)::date IS NULL OR b.date = sqlc.narg(date))
ORDER BY b.date DESC, b.time_start DESC;

-- name: BookingStats :one
SELECT
  COUNT(*) FILTER (WHERE status = 'confirmed') AS confirmed_count,
  COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_count,
  COUNT(*) FILTER (WHERE status = 'completed') AS completed_count,
  COALESCE(SUM(s.price_pence) FILTER (WHERE b.status IN ('confirmed','completed')), 0) AS total_pence
FROM bookings b
JOIN services s ON s.id = b.service_id
WHERE b.date BETWEEN $1 AND $2;
