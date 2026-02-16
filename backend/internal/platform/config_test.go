package platform

import (
	"testing"
)

func TestThatLoadConfigFromEnvDefaultsJWTSecretInDevelopment(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "development")
	t.Setenv("AUTH_MODE", "legacy")
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
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.JWTSecret != "prod-jwt-secret" {
		t.Fatalf("expected JWTSecret to remain prod-jwt-secret, got %q", cfg.JWTSecret)
	}
}

func TestThatLoadConfigFromEnvReturnsErrorWhenJWTSecretMissingInLegacyOutsideDevelopment(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")

	// WHEN
	_, err := LoadConfigFromEnv()

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestThatLoadConfigFromEnvParsesHTTPTimeoutsAndMaxBodyWhenValid(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("HTTP_READ_TIMEOUT_SECONDS", "11")
	t.Setenv("HTTP_WRITE_TIMEOUT_SECONDS", "12")
	t.Setenv("HTTP_IDLE_TIMEOUT_SECONDS", "13")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT_SECONDS", "14")
	t.Setenv("HTTP_MAX_BODY_BYTES", "12345")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.HTTPReadTimeoutSeconds != 11 || cfg.HTTPWriteTimeoutSeconds != 12 || cfg.HTTPIdleTimeoutSeconds != 13 || cfg.HTTPReadHeaderTimeoutSeconds != 14 || cfg.HTTPMaxBodyBytes != 12345 {
		t.Fatalf("expected timeouts/body 11/12/13/14/12345, got %d/%d/%d/%d/%d", cfg.HTTPReadTimeoutSeconds, cfg.HTTPWriteTimeoutSeconds, cfg.HTTPIdleTimeoutSeconds, cfg.HTTPReadHeaderTimeoutSeconds, cfg.HTTPMaxBodyBytes)
	}
}

func TestThatLoadConfigFromEnvUsesDefaultHTTPTimeoutsWhenInvalid(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("HTTP_READ_TIMEOUT_SECONDS", "0")
	t.Setenv("HTTP_WRITE_TIMEOUT_SECONDS", "-1")
	t.Setenv("HTTP_IDLE_TIMEOUT_SECONDS", "abc")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT_SECONDS", "")
	t.Setenv("HTTP_MAX_BODY_BYTES", "0")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.HTTPReadTimeoutSeconds != 15 || cfg.HTTPWriteTimeoutSeconds != 15 || cfg.HTTPIdleTimeoutSeconds != 60 || cfg.HTTPReadHeaderTimeoutSeconds != 5 || cfg.HTTPMaxBodyBytes != 2*1024*1024 {
		t.Fatalf("expected default http settings, got %d/%d/%d/%d/%d", cfg.HTTPReadTimeoutSeconds, cfg.HTTPWriteTimeoutSeconds, cfg.HTTPIdleTimeoutSeconds, cfg.HTTPReadHeaderTimeoutSeconds, cfg.HTTPMaxBodyBytes)
	}
}

func TestThatLoadConfigFromEnvParsesPGXPoolSettingsWhenValid(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("PGX_POOL_MAX_CONNS", "21")
	t.Setenv("PGX_POOL_MIN_CONNS", "2")
	t.Setenv("PGX_POOL_MAX_CONN_LIFETIME_SECONDS", "300")
	t.Setenv("PGX_POOL_HEALTH_CHECK_PERIOD_SECONDS", "7")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.PGXPoolMaxConns != 21 || cfg.PGXPoolMinConns != 2 || cfg.PGXPoolMaxConnLifetimeSeconds != 300 || cfg.PGXPoolHealthCheckPeriodSeconds != 7 {
		t.Fatalf("expected pgx pool settings 21/2/300/7, got %d/%d/%d/%d", cfg.PGXPoolMaxConns, cfg.PGXPoolMinConns, cfg.PGXPoolMaxConnLifetimeSeconds, cfg.PGXPoolHealthCheckPeriodSeconds)
	}
}

func TestThatLoadConfigFromEnvReturnsErrorWhenPGXPoolMinExceedsMax(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("PGX_POOL_MAX_CONNS", "5")
	t.Setenv("PGX_POOL_MIN_CONNS", "6")

	// WHEN
	_, err := LoadConfigFromEnv()

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestThatLoadConfigFromEnvParsesRateLimitRPMWhenValid(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("RATE_LIMIT_RPM", "123")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.RateLimitRPM != 123 {
		t.Fatalf("expected RateLimitRPM 123, got %d", cfg.RateLimitRPM)
	}
}

func TestThatLoadConfigFromEnvAllowsMetricsWithoutTokenInDevelopment(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "development")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_AUTH_TOKEN", "")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cfg.MetricsEnabled {
		t.Fatalf("expected MetricsEnabled true")
	}
}

func TestThatLoadConfigFromEnvReturnsErrorWhenMetricsEnabledInProductionWithoutToken(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("METRICS_ENABLED", "true")
	t.Setenv("METRICS_AUTH_TOKEN", "")

	// WHEN
	_, err := LoadConfigFromEnv()

	// THEN
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestThatLoadConfigFromEnvDoesNotRequireJWTSecretInCognitoMode(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "cognito")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("COGNITO_REGION", "us-east-1")
	t.Setenv("COGNITO_USER_POOL_ID", "us-east-1_abc123")
	t.Setenv("COGNITO_APP_CLIENT_ID", "client123")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.AuthMode != "cognito" {
		t.Fatalf("expected AuthMode cognito, got %q", cfg.AuthMode)
	}
}

func TestThatLoadConfigFromEnvRequiresCognitoEnvVarsInCognitoMode(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "cognito")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("COGNITO_REGION", "")
	t.Setenv("COGNITO_USER_POOL_ID", "")
	t.Setenv("COGNITO_APP_CLIENT_ID", "")

	// WHEN
	_, err := LoadConfigFromEnv()

	// THEN
	if err == nil {
		t.Fatalf("expected error")
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
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("PORT", "")
	t.Setenv("NODE_ENV", "development")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "")

	// WHEN
	cfg, err := LoadConfigFromEnv()

	// THEN
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.AllowedOrigins) == 0 || cfg.AllowedOrigins[0] != "http://localhost:3000" || cfg.Port != "8080" {
		t.Fatalf("expected defaults origin/port, got %v/%q", cfg.AllowedOrigins, cfg.Port)
	}
}

func TestThatLoadConfigFromEnvReturnsErrorWhenCorsAllowlistMissingInProduction(t *testing.T) {
	// GIVEN
	t.Setenv("NODE_ENV", "production")
	t.Setenv("AUTH_MODE", "legacy")
	t.Setenv("JWT_SECRET", "prod-jwt-secret")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGIN", "")
	t.Setenv("ALLOWED_ORIGINS", "")

	// WHEN
	_, err := LoadConfigFromEnv()

	// THEN
	if err == nil {
		t.Fatalf("expected error")
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

func TestThatDevModeIsRejectedOutsideDevelopment(t *testing.T) {
	// GIVEN AUTH_MODE=dev and NODE_ENV=production
	t.Setenv("AUTH_MODE", "dev")
	t.Setenv("NODE_ENV", "production")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DOTENV_ENABLED", "false")

	// WHEN loading config
	_, err := LoadConfigFromEnv()

	// THEN an error is returned
	if err == nil {
		t.Error("expected error when AUTH_MODE=dev outside development")
	}
}

func TestThatDevModeIsAllowedInDevelopment(t *testing.T) {
	// GIVEN AUTH_MODE=dev and NODE_ENV=development
	t.Setenv("AUTH_MODE", "dev")
	t.Setenv("NODE_ENV", "development")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/testdb")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("DOTENV_ENABLED", "false")

	// WHEN loading config
	cfg, err := LoadConfigFromEnv()

	// THEN no error is returned and auth mode is dev
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.AuthMode != "dev" {
		t.Errorf("expected AuthMode=dev, got %q", cfg.AuthMode)
	}
}
