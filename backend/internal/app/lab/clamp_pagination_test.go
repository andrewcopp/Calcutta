package lab

import "testing"

func TestThatClampPaginationDefaultsLimitWhenZero(t *testing.T) {
	// GIVEN a limit of zero
	limit, offset := 0, 0

	// WHEN clamping pagination
	gotLimit, _ := clampPagination(limit, offset)

	// THEN the limit defaults to 50
	if gotLimit != 50 {
		t.Errorf("expected limit 50, got %d", gotLimit)
	}
}

func TestThatClampPaginationDefaultsLimitWhenNegative(t *testing.T) {
	// GIVEN a negative limit
	limit, offset := -5, 0

	// WHEN clamping pagination
	gotLimit, _ := clampPagination(limit, offset)

	// THEN the limit defaults to 50
	if gotLimit != 50 {
		t.Errorf("expected limit 50, got %d", gotLimit)
	}
}

func TestThatClampPaginationCapsLimitWhenTooHigh(t *testing.T) {
	// GIVEN a limit exceeding the maximum of 200
	limit, offset := 500, 0

	// WHEN clamping pagination
	gotLimit, _ := clampPagination(limit, offset)

	// THEN the limit is capped at 200
	if gotLimit != 200 {
		t.Errorf("expected limit 200, got %d", gotLimit)
	}
}

func TestThatClampPaginationKeepsValidLimit(t *testing.T) {
	// GIVEN a valid limit within bounds
	limit, offset := 75, 0

	// WHEN clamping pagination
	gotLimit, _ := clampPagination(limit, offset)

	// THEN the limit is unchanged
	if gotLimit != 75 {
		t.Errorf("expected limit 75, got %d", gotLimit)
	}
}

func TestThatClampPaginationDefaultsNegativeOffset(t *testing.T) {
	// GIVEN a negative offset
	limit, offset := 50, -10

	// WHEN clamping pagination
	_, gotOffset := clampPagination(limit, offset)

	// THEN the offset defaults to 0
	if gotOffset != 0 {
		t.Errorf("expected offset 0, got %d", gotOffset)
	}
}

func TestThatClampPaginationKeepsValidOffset(t *testing.T) {
	// GIVEN a valid non-negative offset
	limit, offset := 50, 25

	// WHEN clamping pagination
	_, gotOffset := clampPagination(limit, offset)

	// THEN the offset is unchanged
	if gotOffset != 25 {
		t.Errorf("expected offset 25, got %d", gotOffset)
	}
}
