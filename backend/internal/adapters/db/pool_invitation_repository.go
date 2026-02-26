package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolInvitationRepository struct {
	q *sqlc.Queries
}

func NewPoolInvitationRepository(pool *pgxpool.Pool) *PoolInvitationRepository {
	return &PoolInvitationRepository{q: sqlc.New(pool)}
}

func (r *PoolInvitationRepository) CreateInvitation(ctx context.Context, invitation *models.PoolInvitation) error {
	invitation.ID = uuid.New().String()
	params := sqlc.CreatePoolInvitationParams{
		ID:        invitation.ID,
		PoolID:    invitation.PoolID,
		UserID:    invitation.UserID,
		InvitedBy: invitation.InvitedBy,
		Status:    "pending",
	}
	if err := r.q.CreatePoolInvitation(ctx, params); err != nil {
		return fmt.Errorf("creating pool invitation: %w", err)
	}
	return nil
}

func (r *PoolInvitationRepository) ListInvitations(ctx context.Context, poolID string) ([]*models.PoolInvitation, error) {
	rows, err := r.q.ListPoolInvitationsByPoolID(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("listing invitations for pool %s: %w", poolID, err)
	}
	out := make([]*models.PoolInvitation, 0, len(rows))
	for _, row := range rows {
		inv := &models.PoolInvitation{
			ID:        row.ID,
			PoolID:    row.PoolID,
			UserID:    row.UserID,
			InvitedBy: row.InvitedBy,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		}
		if row.RevokedAt.Valid {
			t := row.RevokedAt.Time
			inv.RevokedAt = &t
		}
		out = append(out, inv)
	}
	return out, nil
}

func (r *PoolInvitationRepository) GetInvitationByPoolAndUser(ctx context.Context, poolID, userID string) (*models.PoolInvitation, error) {
	row, err := r.q.GetPoolInvitationByPoolAndUser(ctx, sqlc.GetPoolInvitationByPoolAndUserParams{
		PoolID: poolID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "invitation", ID: poolID + "/" + userID}
		}
		return nil, fmt.Errorf("getting invitation for pool %s and user %s: %w", poolID, userID, err)
	}
	inv := &models.PoolInvitation{
		ID:        row.ID,
		PoolID:    row.PoolID,
		UserID:    row.UserID,
		InvitedBy: row.InvitedBy,
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		inv.RevokedAt = &t
	}
	return inv, nil
}

func (r *PoolInvitationRepository) GetPendingInvitationByPoolAndUser(ctx context.Context, poolID, userID string) (*models.PoolInvitation, error) {
	row, err := r.q.GetPendingPoolInvitationByPoolAndUser(ctx, sqlc.GetPendingPoolInvitationByPoolAndUserParams{
		PoolID: poolID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "invitation", ID: poolID + "/" + userID}
		}
		return nil, fmt.Errorf("getting pending invitation for pool %s and user %s: %w", poolID, userID, err)
	}
	inv := &models.PoolInvitation{
		ID:        row.ID,
		PoolID:    row.PoolID,
		UserID:    row.UserID,
		InvitedBy: row.InvitedBy,
		Status:    row.Status,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		inv.RevokedAt = &t
	}
	return inv, nil
}

func (r *PoolInvitationRepository) AcceptInvitation(ctx context.Context, id string) error {
	affected, err := r.q.AcceptPoolInvitation(ctx, id)
	if err != nil {
		return fmt.Errorf("accepting invitation %s: %w", id, err)
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "invitation", ID: id}
	}
	return nil
}

func (r *PoolInvitationRepository) RevokeInvitation(ctx context.Context, id string) error {
	affected, err := r.q.RevokePoolInvitation(ctx, id)
	if err != nil {
		return fmt.Errorf("revoking invitation %s: %w", id, err)
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "invitation", ID: id}
	}
	return nil
}

func (r *PoolInvitationRepository) ListPendingInvitationsByUserID(ctx context.Context, userID string) ([]*models.PoolInvitation, error) {
	rows, err := r.q.ListPendingInvitationsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing pending invitations for user %s: %w", userID, err)
	}
	out := make([]*models.PoolInvitation, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.PoolInvitation{
			ID:        row.ID,
			PoolID:    row.PoolID,
			UserID:    row.UserID,
			InvitedBy: row.InvitedBy,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}
	return out, nil
}
