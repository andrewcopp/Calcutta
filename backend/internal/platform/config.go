package platform

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL               string
	AllowedOrigin             string
	Port                      string
	AuthMode                  string
	JWTSecret                 string
	AccessTokenTTLSeconds     int
	RefreshTokenTTLHours      int
	ShutdownTimeoutSeconds    int
	CognitoRegion             string
	CognitoUserPoolID         string
	CognitoAppClientID        string
	CognitoAutoProvision      bool
	CognitoAllowUnprovisioned bool
}

func envBool(key string, defaultValue bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return defaultValue
	}
	switch v {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

func LoadConfigFromEnv() (Config, error) {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}

	authMode := strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_MODE")))
	if authMode == "" {
		authMode = "legacy"
	}
	if authMode != "legacy" && authMode != "cognito" && authMode != "dev" {
		return Config{}, fmt.Errorf("invalid AUTH_MODE %q (expected legacy, cognito, dev)", authMode)
	}

	accessTTLSeconds := 900
	if v := os.Getenv("ACCESS_TOKEN_TTL_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			accessTTLSeconds = parsed
		}
	}

	refreshTTLHours := 24 * 30
	if v := os.Getenv("REFRESH_TOKEN_TTL_HOURS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			refreshTTLHours = parsed
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		dbUser := os.Getenv("DB_USER")
		dbPassword := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbSSLMode := os.Getenv("DB_SSLMODE")
		if dbSSLMode == "" {
			dbSSLMode = "disable"
		}

		if dbUser != "" && dbName != "" && dbHost != "" && dbPort != "" {
			u := &url.URL{
				Scheme: "postgresql",
				Host:   net.JoinHostPort(dbHost, dbPort),
				Path:   "/" + dbName,
			}
			if dbPassword != "" {
				u.User = url.UserPassword(dbUser, dbPassword)
			} else {
				u.User = url.User(dbUser)
			}
			q := u.Query()
			q.Set("sslmode", dbSSLMode)
			u.RawQuery = q.Encode()
			databaseURL = u.String()
		}
	}

	cfg := Config{
		DatabaseURL:               databaseURL,
		AllowedOrigin:             os.Getenv("ALLOWED_ORIGIN"),
		Port:                      os.Getenv("PORT"),
		AuthMode:                  authMode,
		JWTSecret:                 os.Getenv("JWT_SECRET"),
		AccessTokenTTLSeconds:     accessTTLSeconds,
		RefreshTokenTTLHours:      refreshTTLHours,
		ShutdownTimeoutSeconds:    30,
		CognitoRegion:             os.Getenv("COGNITO_REGION"),
		CognitoUserPoolID:         os.Getenv("COGNITO_USER_POOL_ID"),
		CognitoAppClientID:        os.Getenv("COGNITO_APP_CLIENT_ID"),
		CognitoAutoProvision:      envBool("COGNITO_AUTO_PROVISION", false),
		CognitoAllowUnprovisioned: envBool("COGNITO_ALLOW_UNPROVISIONED", false),
	}

	if env == "development" && cfg.AuthMode != "cognito" && cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-jwt-secret"
	}

	if cfg.AllowedOrigin == "" {
		cfg.AllowedOrigin = "http://localhost:3000"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	if cfg.AuthMode == "cognito" {
		if strings.TrimSpace(cfg.CognitoRegion) == "" {
			return Config{}, fmt.Errorf("COGNITO_REGION environment variable is not set")
		}
		if strings.TrimSpace(cfg.CognitoUserPoolID) == "" {
			return Config{}, fmt.Errorf("COGNITO_USER_POOL_ID environment variable is not set")
		}
		if strings.TrimSpace(cfg.CognitoAppClientID) == "" {
			return Config{}, fmt.Errorf("COGNITO_APP_CLIENT_ID environment variable is not set")
		}
	}
	if cfg.AuthMode != "cognito" && strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	return cfg, nil
}
