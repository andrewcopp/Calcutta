package platform

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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

func envString(key string, defaultValue string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	return v
}

func envBoolPtr(key string) *bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return nil
	}
	var result bool
	switch v {
	case "1", "true", "t", "yes", "y", "on":
		result = true
	case "0", "false", "f", "no", "n", "off":
		result = false
	default:
		return nil
	}
	return &result
}

func loadDotEnvFiles() {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	var foundDir string
	dir := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			foundDir = dir
			break
		}
		if _, err := os.Stat(filepath.Join(dir, ".env.local")); err == nil {
			foundDir = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if foundDir == "" {
		return
	}

	if err := loadDotEnvFile(filepath.Join(foundDir, ".env")); err != nil && !os.IsNotExist(err) {
		slog.Warn("failed to load .env", "path", filepath.Join(foundDir, ".env"), "error", err)
	}
	if err := loadDotEnvFile(filepath.Join(foundDir, ".env.local")); err != nil && !os.IsNotExist(err) {
		slog.Warn("failed to load .env.local", "path", filepath.Join(foundDir, ".env.local"), "error", err)
	}
}

func loadDotEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		if key == "" {
			continue
		}
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, ok := os.LookupEnv(key); ok {
			continue
		}
		_ = os.Setenv(key, val)
	}
	return scanner.Err()
}

func isGoTestProcess() bool {
	if strings.HasSuffix(os.Args[0], ".test") {
		return true
	}
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.") {
			return true
		}
	}
	return false
}

func loadAppEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "production"
	}
	return env
}

func loadAllowedOrigins(env string) []string {
	allowedOriginsEnv := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if allowedOriginsEnv == "" {
		allowedOriginsEnv = strings.TrimSpace(os.Getenv("ALLOWED_ORIGIN"))
	}
	if allowedOriginsEnv == "" && env == "development" {
		allowedOriginsEnv = "http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173"
	}
	allowedOrigins := make([]string, 0)
	for _, o := range strings.Split(allowedOriginsEnv, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			allowedOrigins = append(allowedOrigins, trimmed)
		}
	}
	return allowedOrigins
}

func loadAuthConfig(env string) (authMode string, accessTTLSeconds int, refreshTTLHours int, err error) {
	authMode = strings.ToLower(strings.TrimSpace(os.Getenv("AUTH_MODE")))
	if authMode == "" {
		authMode = "jwt"
	}
	// Support "legacy" as alias for "jwt" for backwards compatibility
	if authMode == "legacy" {
		authMode = "jwt"
	}
	if authMode != "jwt" && authMode != "cognito" && authMode != "dev" {
		return "", 0, 0, fmt.Errorf("invalid AUTH_MODE %q (expected jwt, cognito, dev)", authMode)
	}
	if authMode == "dev" && env != "development" {
		return "", 0, 0, fmt.Errorf("AUTH_MODE=dev is only allowed when APP_ENV=development")
	}

	accessTTLSeconds = 900
	if v := os.Getenv("ACCESS_TOKEN_TTL_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			accessTTLSeconds = parsed
		}
	}

	refreshTTLHours = 24 * 30
	if v := os.Getenv("REFRESH_TOKEN_TTL_HOURS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			refreshTTLHours = parsed
		}
	}

	return authMode, accessTTLSeconds, refreshTTLHours, nil
}

func loadHTTPConfig() (readTimeout, writeTimeout, idleTimeout, readHeaderTimeout, rateLimitRPM int,
	maxBodyBytes int64, metricsEnabled bool, metricsAuthToken string) {
	readTimeout = envInt("HTTP_READ_TIMEOUT_SECONDS", 15, 1)
	writeTimeout = envInt("HTTP_WRITE_TIMEOUT_SECONDS", 15, 1)
	idleTimeout = envInt("HTTP_IDLE_TIMEOUT_SECONDS", 60, 1)
	readHeaderTimeout = envInt("HTTP_READ_HEADER_TIMEOUT_SECONDS", 5, 1)
	maxBodyBytes = envInt64("HTTP_MAX_BODY_BYTES", 2*1024*1024, 1)
	rateLimitRPM = envInt("RATE_LIMIT_RPM", 600, 0)
	metricsEnabled = envBool("METRICS_ENABLED", false)
	metricsAuthToken = strings.TrimSpace(os.Getenv("METRICS_AUTH_TOKEN"))
	return
}

func loadSMTPConfig() (host string, port int, username, password, fromEmail, fromName string,
	startTLS bool, inviteBaseURL string, inviteResendMinSeconds int) {
	host = strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port = envInt("SMTP_PORT", 587, 0)
	username = strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	password = os.Getenv("SMTP_PASSWORD")
	fromEmail = strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL"))
	fromName = strings.TrimSpace(os.Getenv("SMTP_FROM_NAME"))
	startTLS = envBool("SMTP_STARTTLS", true)
	inviteBaseURL = strings.TrimSpace(os.Getenv("INVITE_BASE_URL"))
	inviteResendMinSeconds = envInt("INVITE_RESEND_MIN_SECONDS", 60, 0)
	return
}

func loadPoolConfig() (maxConns, minConns, maxConnLifetimeSeconds, healthCheckPeriodSeconds,
	statementTimeoutMS, lockTimeoutMS int) {
	maxConns = envInt("PGX_POOL_MAX_CONNS", 10, 1)
	minConns = envInt("PGX_POOL_MIN_CONNS", 0, 0)
	maxConnLifetimeSeconds = envInt("PGX_POOL_MAX_CONN_LIFETIME_SECONDS", 1800, 0)
	healthCheckPeriodSeconds = envInt("PGX_POOL_HEALTH_CHECK_PERIOD_SECONDS", 30, 0)
	statementTimeoutMS = envInt("STATEMENT_TIMEOUT_MS", 0, 0)
	lockTimeoutMS = envInt("LOCK_TIMEOUT_MS", 0, 0)
	return
}

func loadDatabaseURL() string {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		return databaseURL
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "require"
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
		return u.String()
	}

	return ""
}

func loadWorkerConfig() (defaultNSims int, excludedEntryName, pythonBin string, runJobsMaxAttempts int, workerID string) {
	defaultNSims = envInt("DEFAULT_N_SIMS", 10000, 1)
	excludedEntryName = strings.TrimSpace(os.Getenv("EXCLUDED_ENTRY_NAME"))
	pythonBin = envString("PYTHON_BIN", "python3")
	runJobsMaxAttempts = envInt("RUN_JOBS_MAX_ATTEMPTS", 5, 1)
	workerID = envString("HOSTNAME", "worker")
	return
}

func loadCookieConfig() (*bool, string) {
	secure := envBoolPtr("COOKIE_SECURE")
	sameSite := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SAMESITE")))
	return secure, sameSite
}

func validateConfig(cfg Config) error {
	if cfg.AppEnv != "development" {
		if len(cfg.AllowedOrigins) == 0 {
			return fmt.Errorf("ALLOWED_ORIGINS or ALLOWED_ORIGIN must be set")
		}
	}

	if cfg.MetricsEnabled && cfg.AppEnv != "development" && strings.TrimSpace(cfg.MetricsAuthToken) == "" {
		return fmt.Errorf("METRICS_AUTH_TOKEN must be set when METRICS_ENABLED=true outside development")
	}

	if cfg.PGXPoolMinConns > cfg.PGXPoolMaxConns {
		return fmt.Errorf("PGX_POOL_MIN_CONNS cannot exceed PGX_POOL_MAX_CONNS")
	}

	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	if cfg.AuthMode == "cognito" {
		if strings.TrimSpace(cfg.CognitoRegion) == "" {
			return fmt.Errorf("COGNITO_REGION environment variable is not set")
		}
		if strings.TrimSpace(cfg.CognitoUserPoolID) == "" {
			return fmt.Errorf("COGNITO_USER_POOL_ID environment variable is not set")
		}
		if strings.TrimSpace(cfg.CognitoAppClientID) == "" {
			return fmt.Errorf("COGNITO_APP_CLIENT_ID environment variable is not set")
		}
	}
	if cfg.AuthMode != "cognito" && strings.TrimSpace(cfg.BootstrapAdminEmail) != "" && strings.TrimSpace(cfg.BootstrapAdminPassword) == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD must be set when BOOTSTRAP_ADMIN_EMAIL is set in jwt/dev auth modes")
	}
	if cfg.AuthMode != "cognito" && strings.TrimSpace(cfg.JWTSecret) == "" {
		return fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	return nil
}

func LoadConfigFromEnv() (Config, error) {
	if !isGoTestProcess() && envBool("DOTENV_ENABLED", true) {
		loadDotEnvFiles()
	}

	env := loadAppEnv()

	authMode, accessTTLSeconds, refreshTTLHours, err := loadAuthConfig(env)
	if err != nil {
		return Config{}, err
	}

	readTimeout, writeTimeout, idleTimeout, readHeaderTimeout, rateLimitRPM,
		maxBodyBytes, metricsEnabled, metricsAuthToken := loadHTTPConfig()

	smtpHost, smtpPort, smtpUsername, smtpPassword, smtpFromEmail, smtpFromName,
		smtpStartTLS, inviteBaseURL, inviteResendMinSeconds := loadSMTPConfig()

	maxConns, minConns, maxConnLifetimeSeconds, healthCheckPeriodSeconds,
		statementTimeoutMS, lockTimeoutMS := loadPoolConfig()

	defaultNSims, excludedEntryName, pythonBin, runJobsMaxAttempts, workerID := loadWorkerConfig()

	cookieSecure, cookieSameSite := loadCookieConfig()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := Config{
		AppEnv:                          env,
		DatabaseURL:                     loadDatabaseURL(),
		AllowedOrigins:                  loadAllowedOrigins(env),
		Port:                            port,
		BootstrapAdminEmail:             strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")),
		BootstrapAdminPassword:          os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"),
		SMTPHost:                        smtpHost,
		SMTPPort:                        smtpPort,
		SMTPUsername:                    smtpUsername,
		SMTPPassword:                    smtpPassword,
		SMTPFromEmail:                   smtpFromEmail,
		SMTPFromName:                    smtpFromName,
		SMTPStartTLS:                    smtpStartTLS,
		InviteBaseURL:                   inviteBaseURL,
		InviteResendMinSeconds:          inviteResendMinSeconds,
		MetricsEnabled:                  metricsEnabled,
		MetricsAuthToken:                metricsAuthToken,
		HTTPReadTimeoutSeconds:          readTimeout,
		HTTPWriteTimeoutSeconds:         writeTimeout,
		HTTPIdleTimeoutSeconds:          idleTimeout,
		HTTPReadHeaderTimeoutSeconds:    readHeaderTimeout,
		HTTPMaxBodyBytes:                maxBodyBytes,
		RateLimitRPM:                    rateLimitRPM,
		PGXPoolMaxConns:                 int32(maxConns),
		PGXPoolMinConns:                 int32(minConns),
		PGXPoolMaxConnLifetimeSeconds:   maxConnLifetimeSeconds,
		PGXPoolHealthCheckPeriodSeconds: healthCheckPeriodSeconds,
		StatementTimeoutMS:              statementTimeoutMS,
		LockTimeoutMS:                   lockTimeoutMS,
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
		DefaultNSims:                    defaultNSims,
		ExcludedEntryName:               excludedEntryName,
		PythonBin:                       pythonBin,
		RunJobsMaxAttempts:              runJobsMaxAttempts,
		WorkerID:                        workerID,
		CookieSecure:                    cookieSecure,
		CookieSameSite:                  cookieSameSite,
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
