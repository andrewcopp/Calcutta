package auth

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"
)

type alwaysErrorReaderForInvite struct{}

func (alwaysErrorReaderForInvite) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}

func TestThatNewInviteTokenFromReaderUsesProvidedBytes(t *testing.T) {
	// GIVEN a deterministic reader with known bytes
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	r := bytes.NewReader(b)

	// WHEN generating an invite token
	got, err := NewInviteTokenFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the token is the base64url encoding of the provided bytes
	want := base64.RawURLEncoding.EncodeToString(b)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestThatNewInviteTokenFromReaderReturnsErrorWhenReaderErrors(t *testing.T) {
	// GIVEN a reader that always errors
	r := alwaysErrorReaderForInvite{}

	// WHEN generating an invite token
	_, err := NewInviteTokenFromReader(r)

	// THEN an error is returned
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestThatHashInviteTokenIsDeterministic(t *testing.T) {
	// GIVEN a known token string
	token := "hello"

	// WHEN hashing the invite token
	got := HashInviteToken(token)

	// THEN the result is the SHA-256 hex digest of "hello"
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestThatHashInviteTokenReturnsDifferentHashForDifferentInputs(t *testing.T) {
	// GIVEN two different tokens
	tokenA := "token-a"
	tokenB := "token-b"

	// WHEN hashing both tokens
	hashA := HashInviteToken(tokenA)
	hashB := HashInviteToken(tokenB)

	// THEN the hashes are different
	if hashA == hashB {
		t.Errorf("expected different hashes for different tokens, both got %q", hashA)
	}
}

func TestThatNewInviteTokenFromReaderProduces43CharacterToken(t *testing.T) {
	// GIVEN a deterministic reader with 32 random bytes
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i + 100)
	}
	r := bytes.NewReader(b)

	// WHEN generating an invite token
	got, err := NewInviteTokenFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the token length matches base64url encoding of 32 bytes (43 chars without padding)
	expectedLen := base64.RawURLEncoding.EncodedLen(32)
	if len(got) != expectedLen {
		t.Errorf("expected token length %d, got %d", expectedLen, len(got))
	}
}
