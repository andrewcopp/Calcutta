package db

import (
	"context"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *CalcuttaRepository) GetPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error) {
	rows, err := r.q.ListCalcuttaPayouts(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaPayout, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaPayout{
			ID:          row.ID,
			CalcuttaID:  row.CalcuttaID,
			Position:    int(row.Position),
			AmountCents: int(row.AmountCents),
			Created:     row.CreatedAt.Time,
			Updated:     row.UpdatedAt.Time,
			Deleted:     nil,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) ReplacePayouts(ctx context.Context, calcuttaID string, payouts []*models.CalcuttaPayout) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	now := time.Now()
	qtx := r.q.WithTx(tx)

	// Soft-delete existing payouts
	_, err = qtx.SoftDeletePayoutsByCalcuttaID(ctx, sqlc.SoftDeletePayoutsByCalcuttaIDParams{
		DeletedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		CalcuttaID: calcuttaID,
	})
	if err != nil {
		return err
	}

	// Insert new payouts
	for _, p := range payouts {
		if p == nil {
			continue
		}
		err = qtx.CreatePayout(ctx, sqlc.CreatePayoutParams{
			ID:          uuid.New().String(),
			CalcuttaID:  calcuttaID,
			Position:    int32(p.Position),
			AmountCents: int32(p.AmountCents),
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
