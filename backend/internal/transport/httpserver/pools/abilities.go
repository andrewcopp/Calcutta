package pools

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func computeAbilities(ctx context.Context, authz policy.AuthorizationChecker, userID string, pool *models.Pool) *dtos.PoolAbilities {
	if userID == "" {
		return nil
	}

	canManage, _ := policy.CanManagePool(ctx, authz, userID, pool)
	canInvite, _ := policy.CanInviteToPool(ctx, authz, userID, pool)

	return &dtos.PoolAbilities{
		CanEditSettings:     canManage.Allowed,
		CanInviteUsers:      canInvite.Allowed,
		CanEditPortfolios:   canManage.Allowed,
		CanManageCoManagers: canManage.Allowed,
	}
}
