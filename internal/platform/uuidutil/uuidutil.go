package uuidutil

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// String returns the canonical string form of a pgtype.UUID, or "" if it is not valid.
func String(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

// Parse converts a UUID string into a pgtype.UUID.
func Parse(s string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}
