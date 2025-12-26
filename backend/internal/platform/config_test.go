package platform

import "testing"

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
