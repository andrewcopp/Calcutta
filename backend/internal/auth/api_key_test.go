package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

type alwaysErrorReaderForAPIKey struct{}

func (alwaysErrorReaderForAPIKey) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestThatNewAPIKeyFromReaderUsesProvidedBytes(t *testing.T) {
	// GIVEN
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	r := bytes.NewReader(b)

	// WHEN
	got, err := NewAPIKeyFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN
	want := base64.RawURLEncoding.EncodeToString(b)
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestThatNewAPIKeyFromReaderReturnsErrorWhenReaderErrors(t *testing.T) {
	// GIVEN
	r := alwaysErrorReaderForAPIKey{}

	// WHEN
	_, err := NewAPIKeyFromReader(r)

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}
