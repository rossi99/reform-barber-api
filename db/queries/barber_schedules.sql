-- name: GetScheduleForBarberDOW :one
SELECT * FROM barber_schedules
WHERE barber_id = $1 AND dow = $2
LIMIT 1;

-- name: ListSchedulesForBarber :many
SELECT * FROM barber_schedules
WHERE barber_id = $1
ORDER BY dow ASC;

-- name: UpsertBarberSchedule :one
INSERT INTO barber_schedules (barber_id, dow, open_time, close_time)
VALUES ($1, $2, $3, $4)
ON CONFLICT (barber_id, dow) DO UPDATE
  SET open_time = EXCLUDED.open_time,
      close_time = EXCLUDED.close_time
RETURNING *;

-- name: DeleteBarberSchedule :exec
DELETE FROM barber_schedules WHERE barber_id = $1 AND dow = $2;
