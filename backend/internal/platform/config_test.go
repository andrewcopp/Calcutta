package platform

import (
	"testing"
)

func TestThatLoadConfigFromEnvDefaultsJWTSecretInDevelopment(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "development")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.JWTSecret != "dev-jwt-secret" {
		t.Fatalf("expected JWTSecret to default to dev-jwt-secret, got %q", cfg.JWTSecret)
	}
}

func TestThatLoadConfigFromEnvDoesNotDefaultJWTSecretOutsideDevelopment(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.JWTSecret != "" {
		t.Fatalf("expected JWTSecret to remain empty, got %q", cfg.JWTSecret)
	}
}

func TestThatLoadConfigFromEnvBuildsDatabaseURLFromDBPartsWithPassword(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "db")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_SSLMODE", "disable")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	got := cfg.DatabaseURL
	want := "postgresql://user:pass@localhost:5432/db?sslmode=disable"
	if got != want {
		t.Fatalf("expected DatabaseURL %q, got %q", want, got)
	}
}

func TestThatLoadConfigFromEnvBuildsDatabaseURLFromDBPartsWithoutPassword(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "db")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_SSLMODE", "disable")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	got := cfg.DatabaseURL
	want := "postgresql://user@localhost:5432/db?sslmode=disable"
	if got != want {
		t.Fatalf("expected DatabaseURL %q, got %q", want, got)
	}
}

func TestThatLoadConfigFromEnvDefaultsSSLModeWhenBuildingDatabaseURL(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "db")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_SSLMODE", "")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	got := cfg.DatabaseURL
	want := "postgresql://user@localhost:5432/db?sslmode=disable"
	if got != want {
		t.Fatalf("expected DatabaseURL %q, got %q", want, got)
	}
}

func TestThatLoadConfigFromEnvParsesTokenTTLsWhenValid(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ACCESS_TOKEN_TTL_SECONDS", "123")
	t.Setenv("REFRESH_TOKEN_TTL_HOURS", "456")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AccessTokenTTLSeconds != 123 || cfg.RefreshTokenTTLHours != 456 {
		t.Fatalf("expected ttls 123/456, got %d/%d", cfg.AccessTokenTTLSeconds, cfg.RefreshTokenTTLHours)
	}
}

func TestThatLoadConfigFromEnvKeepsDefaultTTLsWhenInvalid(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ACCESS_TOKEN_TTL_SECONDS", "-1")
	t.Setenv("REFRESH_TOKEN_TTL_HOURS", "abc")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AccessTokenTTLSeconds != 900 || cfg.RefreshTokenTTLHours != 24*30 {
		t.Fatalf("expected default ttls 900/%d, got %d/%d", 24*30, cfg.AccessTokenTTLSeconds, cfg.RefreshTokenTTLHours)
	}
}

func TestThatLoadConfigFromEnvDefaultsAllowedOriginAndPort(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "")
	t.Setenv("PORT", "")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AllowedOrigin != "http://localhost:3000" || cfg.Port != "8080" {
		t.Fatalf("expected defaults origin/port, got %q/%q", cfg.AllowedOrigin, cfg.Port)
	}
}

func TestThatLoadConfigFromEnvReturnsErrorWhenDatabaseURLCannotBeDetermined(t *testing.T) {
	// GIVEN
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")

	// WHEN
	_, err := LoadConfigFromEnv()

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}
