package lab

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
)

// Handler handles lab-related HTTP requests.
type Handler struct {
	app        *app.App
	authUserID func(context.Context) string
}

// NewHandlerWithAuthUserID creates a new lab handler with auth user ID function.
func NewHandlerWithAuthUserID(a *app.App, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authUserID: authUserID}
}
