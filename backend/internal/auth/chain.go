package auth

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

var _ ports.Authenticator = (*ChainAuthenticator)(nil)

// ChainAuthenticator tries authenticators in order.
// (nil, nil) = skip to next. (nil, err) = stop and return error. (identity, nil) = success.
type ChainAuthenticator struct {
	authenticators []ports.Authenticator
}

func NewChainAuthenticator(authenticators ...ports.Authenticator) *ChainAuthenticator {
	return &ChainAuthenticator{authenticators: authenticators}
}

func (c *ChainAuthenticator) Authenticate(ctx context.Context, token string) (*ports.AuthIdentity, error) {
	for _, a := range c.authenticators {
		identity, err := a.Authenticate(ctx, token)
		if err != nil {
			return nil, err
		}
		if identity != nil {
			return identity, nil
		}
	}
	return nil, nil
}
