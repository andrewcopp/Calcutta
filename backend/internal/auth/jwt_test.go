package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestThatNewTokenManagerRejectsBlankSecret(t *testing.T) {
	// GIVEN
	secret := ""
	ttl := 15 * time.Minute

	// WHEN
	_, err := NewTokenManager(secret, ttl)

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestThatNewTokenManagerRejectsNonPositiveTTL(t *testing.T) {
	// GIVEN
	secret := "s"
	ttl := 0 * time.Second

	// WHEN
	_, err := NewTokenManager(secret, ttl)

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestThatIssueAccessTokenRejectsMissingUserID(t *testing.T) {
	// GIVEN
	mgr, err := NewTokenManager("secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHEN
	_, _, issueErr := mgr.IssueAccessToken("", "sid", time.Unix(10, 0).UTC())

	// THEN
	if issueErr == nil {
		t.Fatalf("expected error")
	}
}

func TestThatIssueAccessTokenRejectsMissingSessionID(t *testing.T) {
	// GIVEN
	mgr, err := NewTokenManager("secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHEN
	_, _, issueErr := mgr.IssueAccessToken("uid", "", time.Unix(10, 0).UTC())

	// THEN
	if issueErr == nil {
		t.Fatalf("expected error")
	}
}

func TestThatIssueAccessTokenAndVerifyAccessTokenRoundTripReturnsClaims(t *testing.T) {
	// GIVEN
	now := time.Unix(1700000000, 0).UTC()
	mgr, err := NewTokenManager("secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tok, _, err := mgr.IssueAccessToken("user-1", "sess-1", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHEN
	claims, err := mgr.VerifyAccessToken(tok, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN
	if claims.Sub != "user-1" {
		t.Fatalf("expected sub to equal user-1, got %q", claims.Sub)
	}
}

func TestThatVerifyAccessTokenRejectsInvalidTokenFormat(t *testing.T) {
	// GIVEN
	mgr, err := NewTokenManager("secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHEN
	_, verifyErr := mgr.VerifyAccessToken("not-a-jwt", time.Unix(10, 0).UTC())

	// THEN
	if verifyErr == nil {
		t.Fatalf("expected error")
	}
}

func TestThatVerifyAccessTokenRejectsExpiredToken(t *testing.T) {
	// GIVEN
	issueNow := time.Unix(1700000000, 0).UTC()
	mgr, err := NewTokenManager("secret", 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok, _, err := mgr.IssueAccessToken("user-1", "sess-1", issueNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHEN
	_, verifyErr := mgr.VerifyAccessToken(tok, issueNow.Add(2*time.Second))

	// THEN
	if verifyErr == nil {
		t.Fatalf("expected error")
	}
}

func TestThatIssuedTokenContainsKidIssAndVerClaims(t *testing.T) {
	// GIVEN a token manager
	mgr, err := NewTokenManager("secret", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := time.Unix(1700000000, 0).UTC()

	// WHEN issuing a token
	tok, _, err := mgr.IssueAccessToken("user-1", "sess-1", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the raw header contains kid="v1" and payload contains iss="calcutta", ver=1
	parts := strings.Split(tok, ".")
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("failed to decode header: %v", err)
	}
	var header map[string]string
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("failed to unmarshal header: %v", err)
	}
	if header["kid"] != "v1" {
		t.Errorf("expected kid=v1, got %q", header["kid"])
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["iss"] != "calcutta" {
		t.Errorf("expected iss=calcutta, got %v", payload["iss"])
	}
	if payload["ver"] != float64(1) {
		t.Errorf("expected ver=1, got %v", payload["ver"])
	}
}

func TestThatVerifyAccessTokenRejectsSignatureMismatch(t *testing.T) {
	// GIVEN
	now := time.Unix(1700000000, 0).UTC()
	mgrA, err := NewTokenManager("secret-a", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mgrB, err := NewTokenManager("secret-b", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okTok, _, err := mgrA.IssueAccessToken("user-1", "sess-1", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// WHEN
	_, verifyErr := mgrB.VerifyAccessToken(okTok, now)

	// THEN
	if verifyErr == nil {
		t.Fatalf("expected error")
	}
}
