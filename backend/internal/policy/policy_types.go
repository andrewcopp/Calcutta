package policy

import "context"

type AuthorizationChecker interface {
	HasPermission(ctx context.Context, userID string, scope string, scopeID string, permission string) (bool, error)
}

type Decision struct {
	Allowed bool
	IsAdmin bool
	Status  int
	Code    string
	Message string
}

const (
	permissionAdminOverride = "pool.config.write"
)
