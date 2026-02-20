package db

import "testing"

func TestThatImportKeySuffixIsSixCharacters(t *testing.T) {
	// GIVEN an arbitrary tournament ID
	id := "some-random-tournament-id"

	// WHEN computing the import key suffix
	suffix := computeImportKeySuffix(id)

	// THEN the suffix is exactly 6 characters
	if len(suffix) != 6 {
		t.Errorf("expected suffix length 6, got %d", len(suffix))
	}
}

func TestThatImportKeySuffixMatchesExpectedMD5(t *testing.T) {
	// GIVEN a known tournament ID
	id := "test-uuid-123"

	// WHEN computing the import key suffix
	suffix := computeImportKeySuffix(id)

	// THEN it matches the first 6 hex chars of md5("test-uuid-123")
	if suffix != "f1b48e" {
		t.Errorf("expected suffix f1b48e, got %s", suffix)
	}
}
