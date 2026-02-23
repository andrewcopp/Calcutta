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

type CalcuttaInvitationRepository struct {
	q *sqlc.Queries
}

func NewCalcuttaInvitationRepository(pool *pgxpool.Pool) *CalcuttaInvitationRepository {
	return &CalcuttaInvitationRepository{q: sqlc.New(pool)}
}

func (r *CalcuttaInvitationRepository) CreateInvitation(ctx context.Context, invitation *models.CalcuttaInvitation) error {
	invitation.ID = uuid.New().String()
	params := sqlc.CreateCalcuttaInvitationParams{
		ID:         invitation.ID,
		CalcuttaID: invitation.CalcuttaID,
		UserID:     invitation.UserID,
		InvitedBy:  invitation.InvitedBy,
		Status:     "pending",
	}
	if err := r.q.CreateCalcuttaInvitation(ctx, params); err != nil {
		return fmt.Errorf("creating calcutta invitation: %w", err)
	}
	return nil
}

func (r *CalcuttaInvitationRepository) ListInvitations(ctx context.Context, calcuttaID string) ([]*models.CalcuttaInvitation, error) {
	rows, err := r.q.ListCalcuttaInvitationsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, fmt.Errorf("listing invitations for calcutta %s: %w", calcuttaID, err)
	}
	out := make([]*models.CalcuttaInvitation, 0, len(rows))
	for _, row := range rows {
		inv := &models.CalcuttaInvitation{
			ID:         row.ID,
			CalcuttaID: row.CalcuttaID,
			UserID:     row.UserID,
			InvitedBy:  row.InvitedBy,
			Status:     row.Status,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		}
		if row.RevokedAt.Valid {
			t := row.RevokedAt.Time
			inv.RevokedAt = &t
		}
		out = append(out, inv)
	}
	return out, nil
}

func (r *CalcuttaInvitationRepository) GetInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error) {
	row, err := r.q.GetCalcuttaInvitationByCalcuttaAndUser(ctx, sqlc.GetCalcuttaInvitationByCalcuttaAndUserParams{
		CalcuttaID: calcuttaID,
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "invitation", ID: calcuttaID + "/" + userID}
		}
		return nil, fmt.Errorf("getting invitation for calcutta %s and user %s: %w", calcuttaID, userID, err)
	}
	inv := &models.CalcuttaInvitation{
		ID:         row.ID,
		CalcuttaID: row.CalcuttaID,
		UserID:     row.UserID,
		InvitedBy:  row.InvitedBy,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		inv.RevokedAt = &t
	}
	return inv, nil
}

func (r *CalcuttaInvitationRepository) GetPendingInvitationByCalcuttaAndUser(ctx context.Context, calcuttaID, userID string) (*models.CalcuttaInvitation, error) {
	row, err := r.q.GetPendingCalcuttaInvitationByCalcuttaAndUser(ctx, sqlc.GetPendingCalcuttaInvitationByCalcuttaAndUserParams{
		CalcuttaID: calcuttaID,
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "invitation", ID: calcuttaID + "/" + userID}
		}
		return nil, fmt.Errorf("getting pending invitation for calcutta %s and user %s: %w", calcuttaID, userID, err)
	}
	inv := &models.CalcuttaInvitation{
		ID:         row.ID,
		CalcuttaID: row.CalcuttaID,
		UserID:     row.UserID,
		InvitedBy:  row.InvitedBy,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		inv.RevokedAt = &t
	}
	return inv, nil
}

func (r *CalcuttaInvitationRepository) AcceptInvitation(ctx context.Context, id string) error {
	affected, err := r.q.AcceptCalcuttaInvitation(ctx, id)
	if err != nil {
		return fmt.Errorf("accepting invitation %s: %w", id, err)
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "invitation", ID: id}
	}
	return nil
}

func (r *CalcuttaInvitationRepository) RevokeInvitation(ctx context.Context, id string) error {
	affected, err := r.q.RevokeCalcuttaInvitation(ctx, id)
	if err != nil {
		return fmt.Errorf("revoking invitation %s: %w", id, err)
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "invitation", ID: id}
	}
	return nil
}

func (r *CalcuttaInvitationRepository) ListPendingInvitationsByUserID(ctx context.Context, userID string) ([]*models.CalcuttaInvitation, error) {
	rows, err := r.q.ListPendingInvitationsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing pending invitations for user %s: %w", userID, err)
	}
	out := make([]*models.CalcuttaInvitation, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaInvitation{
			ID:         row.ID,
			CalcuttaID: row.CalcuttaID,
			UserID:     row.UserID,
			InvitedBy:  row.InvitedBy,
			Status:     row.Status,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		})
	}
	return out, nil
}
