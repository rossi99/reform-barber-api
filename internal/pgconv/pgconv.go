// Package pgconv provides helpers for converting between Go stdlib types
// and pgtype types used by the sqlc-generated database layer.
package pgconv

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func Date(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func Timestamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func UUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func UUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func FromUUID(u pgtype.UUID) (uuid.UUID, bool) {
	if !u.Valid {
		return uuid.Nil, false
	}
	return uuid.UUID(u.Bytes), true
}

func FromDate(d pgtype.Date) time.Time {
	return d.Time
}

func FromTimestamp(ts pgtype.Timestamptz) time.Time {
	return ts.Time
}
