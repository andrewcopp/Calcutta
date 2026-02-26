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
	// GIVEN an 11-character password
	password := "abcdefghijk"

	// WHEN validating the password
	err := ValidatePassword(password)

	// THEN an invalid error is returned
	if err == nil {
		t.Error("expected error for short password")
	}
}

func TestThatTwelveCharPasswordIsAccepted(t *testing.T) {
	// GIVEN a 12-character password
	password := "abcdefghijkl"

	// WHEN validating the password
	err := ValidatePassword(password)

	// THEN no error is returned
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestThatPasswordOverMaxLengthIsRejected(t *testing.T) {
	// GIVEN a 129-character password
	password := ""
	for i := 0; i < 129; i++ {
		password += "a"
	}

	// WHEN validating the password
	err := ValidatePassword(password)

	// THEN an invalid error is returned
	if err == nil {
		t.Error("expected error for password exceeding max length")
	}
}
