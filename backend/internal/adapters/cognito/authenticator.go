package cognito

import (
	"context"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/google/uuid"
)

// tokenVerifier abstracts the Cognito JWT verification for testability.
type tokenVerifier interface {
	Verify(token string, now time.Time) (*Claims, error)
}

var _ ports.Authenticator = (*Authenticator)(nil)

// Authenticator validates Cognito JWT tokens and resolves users.
type Authenticator struct {
	verifier           tokenVerifier
	users              ports.UserRepository
	autoProvision      bool
	allowUnprovisioned bool
	now                func() time.Time
}

func NewAuthenticator(verifier tokenVerifier, users ports.UserRepository, autoProvision, allowUnprovisioned bool) *Authenticator {
	return &Authenticator{
		verifier:           verifier,
		users:              users,
		autoProvision:      autoProvision,
		allowUnprovisioned: allowUnprovisioned,
		now:                time.Now,
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, token string) (*ports.AuthIdentity, error) {
	claims, err := a.verifier.Verify(token, a.now())
	if err != nil {
		return nil, nil
	}

	cognitoSub := strings.TrimSpace(claims.Sub)
	if cognitoSub == "" {
		return nil, nil
	}

	user, err := a.users.GetByExternalProvider(ctx, "cognito", cognitoSub)
	if err != nil {
		return nil, err
	}

	if user == nil && a.autoProvision {
		user, err = a.provisionUser(ctx, claims, cognitoSub)
		if err != nil {
			return nil, err
		}
	}

	if user != nil {
		if user.Status == "active" {
			return &ports.AuthIdentity{UserID: user.ID}, nil
		}
		return nil, nil
	}

	if a.allowUnprovisioned {
		// Fall back to using cognito sub as user ID directly.
		u, err := a.users.GetByID(ctx, claims.Sub)
		if err != nil {
			return nil, err
		}
		if u != nil && u.Status == "active" {
			return &ports.AuthIdentity{UserID: claims.Sub}, nil
		}
		return nil, nil
	}

	return nil, nil
}

func (a *Authenticator) provisionUser(ctx context.Context, claims *Claims, cognitoSub string) (*models.User, error) {
	id := cognitoSub
	if _, err := uuid.Parse(id); err != nil {
		id = uuid.NewString()
	}

	email := strings.TrimSpace(claims.Email)
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	first := strings.TrimSpace(claims.GivenName)
	last := strings.TrimSpace(claims.FamilyName)
	if first == "" {
		first = "User"
	}
	if last == "" {
		last = "User"
	}

	provider := "cognito"
	user := &models.User{
		ID:                 id,
		Email:              emailPtr,
		FirstName:          first,
		LastName:           last,
		Status:             "active",
		ExternalProvider:   &provider,
		ExternalProviderID: &cognitoSub,
	}
	if err := a.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
