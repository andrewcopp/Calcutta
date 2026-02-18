package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// TimestamptzToPtrTime converts a pgtype.Timestamptz to a *time.Time.
// Returns nil if the timestamp is not valid.
func TimestamptzToPtrTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

// UUIDToPtrString converts a pgtype.UUID to a *string.
// Returns nil if the UUID is not valid.
func UUIDToPtrString(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	s := id.String()
	return &s
}

// TimestamptzToPtrTimeUTC converts a pgtype.Timestamptz to a *time.Time in UTC.
// Returns nil if the timestamp is not valid or is zero.
func TimestamptzToPtrTimeUTC(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	if t.IsZero() {
		return nil
	}
	ut := t.UTC()
	return &ut
}

// DerefString safely dereferences a *string, returning empty string if nil.
func DerefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
