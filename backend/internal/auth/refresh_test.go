package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

type alwaysErrorReader struct{}

func (alwaysErrorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

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

func TestThatNewRefreshTokenFromReaderUsesProvidedBytes(t *testing.T) {
	// GIVEN
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	r := bytes.NewReader(b)

	// WHEN
	got, err := NewRefreshTokenFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN
	want := base64.RawURLEncoding.EncodeToString(b)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestThatNewRefreshTokenFromReaderReturnsErrorWhenReaderErrors(t *testing.T) {
	// GIVEN
	r := alwaysErrorReader{}

	// WHEN
	_, err := NewRefreshTokenFromReader(r)

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}
