package dtos

import "testing"

func TestThatEmptyPasswordIsRejected(t *testing.T) {
	// GIVEN an empty password
	password := ""

	// WHEN validating the password
	err := ValidatePassword(password)

	// THEN a required error is returned
	if err == nil {
		t.Error("expected error for empty password")
	}
}

func TestThatShortPasswordIsRejected(t *testing.T) {
	// GIVEN a 7-character password
	password := "abcdefg"

	// WHEN validating the password
	err := ValidatePassword(password)

	// THEN an invalid error is returned
	if err == nil {
		t.Error("expected error for short password")
	}
}

func TestThatEightCharPasswordIsAccepted(t *testing.T) {
	// GIVEN an 8-character password
	password := "abcdefgh"

	// WHEN validating the password
	err := ValidatePassword(password)

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
