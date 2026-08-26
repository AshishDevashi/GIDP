package pgnum

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

// Int2From converts an int16 into pgtype.Int2, treating 0 as NULL (unset).
func Int2From(v int16) pgtype.Int2 {
	return pgtype.Int2{Int16: v, Valid: v != 0}
}

// Int2To returns the plain int16 value of a pgtype.Int2, or 0 if it is NULL.
func Int2To(t pgtype.Int2) int16 {
	return t.Int16
}

// Int4From converts an int32 into pgtype.Int4, treating 0 as NULL (unset).
func Int4From(v int32) pgtype.Int4 {
	return pgtype.Int4{Int32: v, Valid: v != 0}
}

// Int4To returns the plain int32 value of a pgtype.Int4, or 0 if it is NULL.
func Int4To(t pgtype.Int4) int32 {
	return t.Int32
}

// NumericFrom converts a float64 into pgtype.Numeric, treating 0 as NULL (unset).
func NumericFrom(v float64) pgtype.Numeric {
	if v == 0 {
		return pgtype.Numeric{}
	}
	var n pgtype.Numeric
	_ = n.ScanScientific(strconv.FormatFloat(v, 'f', -1, 64))
	return n
}

// NumericTo returns the plain float64 value of a pgtype.Numeric, or 0 if it is NULL.
func NumericTo(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return 0
	}
	return f.Float64
}
