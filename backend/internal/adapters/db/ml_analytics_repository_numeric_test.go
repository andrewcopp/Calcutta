package db

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func mustPgNumericFromScientific(t *testing.T, s string) pgtype.Numeric {
	t.Helper()

	var n pgtype.Numeric
	if err := n.ScanScientific(s); err != nil {
		t.Fatalf("failed to parse pg numeric from %q: %v", s, err)
	}
	return n
}

func overflowPgNumeric() pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(1), Exp: 10000, Valid: true}
}

func TestThatFloatPtrFromPgNumericReturnsNilWhenNumericIsInvalid(t *testing.T) {
	// GIVEN
	n := pgtype.Numeric{Valid: false}

	// WHEN
	got := floatPtrFromPgNumeric(n)

	// THEN
	if got != nil {
		t.Fatalf("expected nil")
	}
}

func TestThatFloatFromPgNumericReturnsZeroWhenNumericIsInvalid(t *testing.T) {
	// GIVEN
	n := pgtype.Numeric{Valid: false}

	// WHEN
	got := floatFromPgNumeric(n)

	// THEN
	if got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestThatFloatPtrFromPgNumericReturnsNonNilWhenNumericIsValid(t *testing.T) {
	// GIVEN
	n := mustPgNumericFromScientific(t, "123")

	// WHEN
	got := floatPtrFromPgNumeric(n)

	// THEN
	if got == nil {
		t.Fatalf("expected non-nil")
	}
}

func TestThatFloatFromPgNumericReturnsValueWhenNumericIsValid(t *testing.T) {
	// GIVEN
	n := mustPgNumericFromScientific(t, "123")

	// WHEN
	got := floatFromPgNumeric(n)

	// THEN
	if got != 123 {
		t.Fatalf("expected 123, got %v", got)
	}
}

func TestThatFloatPtrFromPgNumericReturnsNilWhenFloat64ConversionOverflows(t *testing.T) {
	// GIVEN
	n := overflowPgNumeric()

	// WHEN
	got := floatPtrFromPgNumeric(n)

	// THEN
	if got != nil {
		t.Fatalf("expected nil")
	}
}

func TestThatFloatFromPgNumericReturnsZeroWhenFloat64ConversionOverflows(t *testing.T) {
	// GIVEN
	n := overflowPgNumeric()

	// WHEN
	got := floatFromPgNumeric(n)

	// THEN
	if got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}
