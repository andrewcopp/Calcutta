package db

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("listing payouts for calcutta %s: %w", calcuttaID, err)
	}

	out := make([]*models.CalcuttaPayout, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaPayout{
			ID:          row.ID,
			CalcuttaID:  row.CalcuttaID,
			Position:    int(row.Position),
			AmountCents: int(row.AmountCents),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			DeletedAt:   nil,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) ReplacePayouts(ctx context.Context, calcuttaID string, payouts []*models.CalcuttaPayout) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to replace payouts for calcutta %s: %w", calcuttaID, err)
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
		return fmt.Errorf("soft-deleting payouts for calcutta %s: %w", calcuttaID, err)
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
			return fmt.Errorf("creating payout for calcutta %s: %w", calcuttaID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to replace payouts for calcutta %s: %w", calcuttaID, err)
	}
	return nil
}
