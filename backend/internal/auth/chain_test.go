package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

// --- stubs ---

type stubAuthenticator struct {
	identity *ports.AuthIdentity
	err      error
}

func (s *stubAuthenticator) Authenticate(_ context.Context, _ string) (*ports.AuthIdentity, error) {
	return s.identity, s.err
}

// --- tests ---

func TestThatChainReturnsFirstSuccessfulMatch(t *testing.T) {
	// GIVEN two authenticators, the first succeeds
	first := &stubAuthenticator{identity: &ports.AuthIdentity{UserID: "u1"}}
	second := &stubAuthenticator{identity: &ports.AuthIdentity{UserID: "u2"}}
	chain := auth.NewChainAuthenticator(first, second)

	// WHEN authenticating
	identity, err := chain.Authenticate(context.Background(), "tok")

	// THEN it returns the first match
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.UserID != "u1" {
		t.Errorf("expected UserID u1, got %s", identity.UserID)
	}
}

func TestThatChainSkipsUnrecognizedTokens(t *testing.T) {
	// GIVEN the first authenticator returns nil,nil and the second succeeds
	first := &stubAuthenticator{}
	second := &stubAuthenticator{identity: &ports.AuthIdentity{UserID: "u2"}}
	chain := auth.NewChainAuthenticator(first, second)

	// WHEN authenticating
	identity, err := chain.Authenticate(context.Background(), "tok")

	// THEN it returns the second match
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.UserID != "u2" {
		t.Errorf("expected UserID u2, got %s", identity.UserID)
	}
}

func TestThatChainStopsOnError(t *testing.T) {
	// GIVEN the first authenticator returns an error
	first := &stubAuthenticator{err: errors.New("recognized but failed")}
	second := &stubAuthenticator{identity: &ports.AuthIdentity{UserID: "u2"}}
	chain := auth.NewChainAuthenticator(first, second)

	// WHEN authenticating
	identity, err := chain.Authenticate(context.Background(), "tok")

	// THEN it stops and returns the error
	if identity != nil {
		t.Errorf("expected nil identity, got %v", identity)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestThatChainReturnsNilNilWhenNoneMatch(t *testing.T) {
	// GIVEN all authenticators return nil,nil
	first := &stubAuthenticator{}
	second := &stubAuthenticator{}
	chain := auth.NewChainAuthenticator(first, second)

	// WHEN authenticating
	identity, err := chain.Authenticate(context.Background(), "tok")

	// THEN it returns nil, nil
	if identity != nil || err != nil {
		t.Errorf("expected (nil, nil), got (%v, %v)", identity, err)
	}
}

// Verify interface compliance at compile time.
var _ ports.Authenticator = (*auth.ChainAuthenticator)(nil)
