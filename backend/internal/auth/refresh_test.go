package auth

import "testing"

func TestThatHashRefreshTokenIsDeterministic(t *testing.T) {
	// GIVEN
	token := "hello"

	// WHEN
	got := HashRefreshToken(token)

	// THEN
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
