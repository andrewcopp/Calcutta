package db

import (
	"context"
	"errors"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
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
	return r.q.CreateCalcuttaInvitation(ctx, params)
}

func (r *CalcuttaInvitationRepository) ListInvitations(ctx context.Context, calcuttaID string) ([]*models.CalcuttaInvitation, error) {
	rows, err := r.q.ListCalcuttaInvitationsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}
	out := make([]*models.CalcuttaInvitation, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaInvitation{
			ID:         row.ID,
			CalcuttaID: row.CalcuttaID,
			UserID:     row.UserID,
			InvitedBy:  row.InvitedBy,
			Status:     row.Status,
			Created:    row.CreatedAt.Time,
			Updated:    row.UpdatedAt.Time,
		})
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
		return nil, err
	}
	return &models.CalcuttaInvitation{
		ID:         row.ID,
		CalcuttaID: row.CalcuttaID,
		UserID:     row.UserID,
		InvitedBy:  row.InvitedBy,
		Status:     row.Status,
		Created:    row.CreatedAt.Time,
		Updated:    row.UpdatedAt.Time,
	}, nil
}

func (r *CalcuttaInvitationRepository) AcceptInvitation(ctx context.Context, id string) error {
	affected, err := r.q.AcceptCalcuttaInvitation(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "invitation", ID: id}
	}
	return nil
}
