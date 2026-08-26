package pgtext

import "github.com/jackc/pgx/v5/pgtype"

// From converts a plain string into pgtype.Text, treating "" as NULL.
func From(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// To returns the plain string value of a pgtype.Text, or "" if it is NULL.
func To(t pgtype.Text) string {
	return t.String
}
