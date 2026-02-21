package platform

type Config struct {
	DatabaseURL                     string
	AllowedOrigins                  []string
	Port                            string
	BootstrapAdminEmail             string
	BootstrapAdminPassword          string
	SMTPHost                        string
	SMTPPort                        int
	SMTPUsername                    string
	SMTPPassword                    string
	SMTPFromEmail                   string
	SMTPFromName                    string
	SMTPStartTLS                    bool
	InviteBaseURL                   string
	InviteResendMinSeconds          int
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

	// Database query safety (0 = no limit)
	StatementTimeoutMS int
	LockTimeoutMS      int

	// Lab/Worker settings
	DefaultNSims       int
	ExcludedEntryName  string
	PythonBin          string
	RunJobsMaxAttempts int
	WorkerID           string

	// Environment
	AppEnv string // "development" or "production"

	// Cookie settings
	CookieSecure   *bool  // nil = auto (secure in production)
	CookieSameSite string // "none", "lax", "strict", or "" for auto
}
