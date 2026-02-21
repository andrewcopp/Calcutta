package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestThatExtractBearerTokenReturnsEmptyStringWhenAuthorizationHeaderIsMissing(t *testing.T) {
	// GIVEN a request with no Authorization header
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the result is an empty string
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestThatExtractBearerTokenReturnsEmptyStringWhenHeaderHasNoSpace(t *testing.T) {
	// GIVEN a request with an Authorization header that has no space separator
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearertoken123")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the result is an empty string
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestThatExtractBearerTokenReturnsEmptyStringWhenSchemeIsNotBearer(t *testing.T) {
	// GIVEN a request with a Basic authorization scheme
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the result is an empty string
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestThatExtractBearerTokenReturnsTokenForValidBearerHeader(t *testing.T) {
	// GIVEN a request with a valid Bearer authorization header
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer token123")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the token is returned
	if got != "token123" {
		t.Errorf("expected %q, got %q", "token123", got)
	}
}

func TestThatExtractBearerTokenIsCaseInsensitiveForBearerScheme(t *testing.T) {
	// GIVEN a request with a lowercase "bearer" scheme
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "bearer token123")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the token is returned
	if got != "token123" {
		t.Errorf("expected %q, got %q", "token123", got)
	}
}

func TestThatExtractBearerTokenIsCaseInsensitiveForMixedCaseScheme(t *testing.T) {
	// GIVEN a request with a mixed-case "BEARER" scheme
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "BEARER token123")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the token is returned
	if got != "token123" {
		t.Errorf("expected %q, got %q", "token123", got)
	}
}

func TestThatExtractBearerTokenTrimsWhitespaceFromToken(t *testing.T) {
	// GIVEN a request with a Bearer token that has leading and trailing whitespace
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer   token123   ")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the token is trimmed
	if got != "token123" {
		t.Errorf("expected %q, got %q", "token123", got)
	}
}

func TestThatExtractBearerTokenReturnsEmptyStringWhenHeaderIsEmptyString(t *testing.T) {
	// GIVEN a request with an empty Authorization header value
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the result is an empty string
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestThatExtractBearerTokenReturnsEmptyStringWhenTokenPartIsOnlyWhitespace(t *testing.T) {
	// GIVEN a request with a Bearer scheme but only whitespace as the token
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer    ")

	// WHEN extracting the bearer token
	got := extractBearerToken(r)

	// THEN the result is an empty string after trimming
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
