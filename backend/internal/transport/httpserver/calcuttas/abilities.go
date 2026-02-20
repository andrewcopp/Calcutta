package calcuttas

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func computeAbilities(ctx context.Context, authz policy.AuthorizationChecker, userID string, calcutta *models.Calcutta) *dtos.CalcuttaAbilities {
	if userID == "" {
		return nil
	}

	canManage, _ := policy.CanManageCalcutta(ctx, authz, userID, calcutta)
	canInvite, _ := policy.CanInviteToCalcutta(ctx, authz, userID, calcutta)

	return &dtos.CalcuttaAbilities{
		CanEditSettings:     canManage.Allowed,
		CanInviteUsers:      canInvite.Allowed,
		CanEditEntries:      canManage.Allowed,
		CanManageCoManagers: canManage.Allowed,
	}
}
