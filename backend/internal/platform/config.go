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
	DatabaseURL                     string
	AllowedOrigins                  []string
	AllowedOrigin                   string
	Port                            string
	MetricsEnabled                  bool
	MetricsAuthToken                string
	HTTPReadTimeoutSeconds          int
	HTTPWriteTimeoutSeconds         int
	HTTPIdleTimeoutSeconds          int
	HTTPReadHeaderTimeoutSeconds    int
	HTTPMaxBodyBytes                int64
	RateLimitRPM                    int
	PGXPoolMaxConns                 int32
	PGXPoolMinConns                 int32
	PGXPoolMaxConnLifetimeSeconds   int
	PGXPoolHealthCheckPeriodSeconds int
	AuthMode                        string
	JWTSecret                       string
	AccessTokenTTLSeconds           int
	RefreshTokenTTLHours            int
	ShutdownTimeoutSeconds          int
	CognitoRegion                   string
	CognitoUserPoolID               string
	CognitoAppClientID              string
	CognitoAutoProvision            bool
	CognitoAllowUnprovisioned       bool
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

func envInt(key string, defaultValue int, minValue int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	if parsed < minValue {
		return defaultValue
	}
	return parsed
}

func envInt64(key string, defaultValue int64, minValue int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultValue
	}
	if parsed < minValue {
		return defaultValue
	}
	return parsed
}

func LoadConfigFromEnv() (Config, error) {
	env := os.Getenv("NODE_ENV")
	if env == "" {
		env = "development"
	}

	allowedOriginsEnv := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if allowedOriginsEnv == "" {
		allowedOriginsEnv = strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
	}
	if allowedOriginsEnv == "" && env == "development" {
		allowedOriginsEnv = "http://localhost:3000"
	}
	allowedOrigins := make([]string, 0)
	for _, o := range strings.Split(allowedOriginsEnv, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
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

	httpReadTimeoutSeconds := envInt("HTTP_READ_TIMEOUT_SECONDS", 15, 1)
	httpWriteTimeoutSeconds := envInt("HTTP_WRITE_TIMEOUT_SECONDS", 15, 1)
	httpIdleTimeoutSeconds := envInt("HTTP_IDLE_TIMEOUT_SECONDS", 60, 1)
	httpReadHeaderTimeoutSeconds := envInt("HTTP_READ_HEADER_TIMEOUT_SECONDS", 5, 1)
	httpMaxBodyBytes := envInt64("HTTP_MAX_BODY_BYTES", 2*1024*1024, 1)
	rateLimitRPM := envInt("RATE_LIMIT_RPM", 300, 0)
	metricsEnabled := envBool("METRICS_ENABLED", false)
	metricsAuthToken := strings.TrimSpace(os.Getenv("METRICS_AUTH_TOKEN"))

	pgxPoolMaxConns := envInt("PGX_POOL_MAX_CONNS", 10, 1)
	pgxPoolMinConns := envInt("PGX_POOL_MIN_CONNS", 0, 0)
	pgxPoolMaxConnLifetimeSeconds := envInt("PGX_POOL_MAX_CONN_LIFETIME_SECONDS", 1800, 0)
	pgxPoolHealthCheckPeriodSeconds := envInt("PGX_POOL_HEALTH_CHECK_PERIOD_SECONDS", 30, 0)

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
		DatabaseURL:                     databaseURL,
		AllowedOrigins:                  allowedOrigins,
		AllowedOrigin:                   os.Getenv("ALLOWED_ORIGIN"),
		Port:                            os.Getenv("PORT"),
		MetricsEnabled:                  metricsEnabled,
		MetricsAuthToken:                metricsAuthToken,
		HTTPReadTimeoutSeconds:          httpReadTimeoutSeconds,
		HTTPWriteTimeoutSeconds:         httpWriteTimeoutSeconds,
		HTTPIdleTimeoutSeconds:          httpIdleTimeoutSeconds,
		HTTPReadHeaderTimeoutSeconds:    httpReadHeaderTimeoutSeconds,
		HTTPMaxBodyBytes:                httpMaxBodyBytes,
		RateLimitRPM:                    rateLimitRPM,
		PGXPoolMaxConns:                 int32(pgxPoolMaxConns),
		PGXPoolMinConns:                 int32(pgxPoolMinConns),
		PGXPoolMaxConnLifetimeSeconds:   pgxPoolMaxConnLifetimeSeconds,
		PGXPoolHealthCheckPeriodSeconds: pgxPoolHealthCheckPeriodSeconds,
		AuthMode:                        authMode,
		JWTSecret:                       os.Getenv("JWT_SECRET"),
		AccessTokenTTLSeconds:           accessTTLSeconds,
		RefreshTokenTTLHours:            refreshTTLHours,
		ShutdownTimeoutSeconds:          30,
		CognitoRegion:                   os.Getenv("COGNITO_REGION"),
		CognitoUserPoolID:               os.Getenv("COGNITO_USER_POOL_ID"),
		CognitoAppClientID:              os.Getenv("COGNITO_APP_CLIENT_ID"),
		CognitoAutoProvision:            envBool("COGNITO_AUTO_PROVISION", false),
		CognitoAllowUnprovisioned:       envBool("COGNITO_ALLOW_UNPROVISIONED", false),
	}
	if len(cfg.AllowedOrigins) > 0 {
		cfg.AllowedOrigin = cfg.AllowedOrigins[0]
	}

	if env == "development" && cfg.AuthMode != "cognito" && cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-jwt-secret"
	}

	if env != "development" {
		if len(cfg.AllowedOrigins) == 0 {
			return Config{}, fmt.Errorf("ALLOWED_ORIGINS or ALLOWED_ORIGIN must be set")
		}
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	if cfg.MetricsEnabled && env != "development" && strings.TrimSpace(cfg.MetricsAuthToken) == "" {
		return Config{}, fmt.Errorf("METRICS_AUTH_TOKEN must be set when METRICS_ENABLED=true outside development")
	}

	if cfg.PGXPoolMinConns > cfg.PGXPoolMaxConns {
		return Config{}, fmt.Errorf("PGX_POOL_MIN_CONNS cannot exceed PGX_POOL_MAX_CONNS")
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
