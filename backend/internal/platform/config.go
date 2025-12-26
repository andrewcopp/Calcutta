package platform

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL            string
	AllowedOrigin          string
	Port                   string
	JWTSecret              string
	AccessTokenTTLSeconds  int
	RefreshTokenTTLHours   int
	ShutdownTimeoutSeconds int
}

func LoadConfigFromEnv() (Config, error) {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
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
		DatabaseURL:            databaseURL,
		AllowedOrigin:          os.Getenv("ALLOWED_ORIGIN"),
		Port:                   os.Getenv("PORT"),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		AccessTokenTTLSeconds:  accessTTLSeconds,
		RefreshTokenTTLHours:   refreshTTLHours,
		ShutdownTimeoutSeconds: 30,
	}

	if env == "development" && cfg.JWTSecret == "" {
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

	return cfg, nil
}
