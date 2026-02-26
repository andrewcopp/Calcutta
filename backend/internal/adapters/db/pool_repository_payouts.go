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

func (r *PoolRepository) GetPayouts(ctx context.Context, poolID string) ([]*models.PoolPayout, error) {
	rows, err := r.q.ListPoolPayouts(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("listing payouts for pool %s: %w", poolID, err)
	}

	out := make([]*models.PoolPayout, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.PoolPayout{
			ID:          row.ID,
			PoolID:      row.PoolID,
			Position:    int(row.Position),
			AmountCents: int(row.AmountCents),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			DeletedAt:   nil,
		})
	}
	return out, nil
}

func (r *PoolRepository) ReplacePayouts(ctx context.Context, poolID string, payouts []*models.PoolPayout) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to replace payouts for pool %s: %w", poolID, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	now := time.Now()
	qtx := r.q.WithTx(tx)

	// Soft-delete existing payouts
	_, err = qtx.SoftDeletePayoutsByPoolID(ctx, sqlc.SoftDeletePayoutsByPoolIDParams{
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		PoolID:    poolID,
	})
	if err != nil {
		return fmt.Errorf("soft-deleting payouts for pool %s: %w", poolID, err)
	}

	// Insert new payouts
	for _, p := range payouts {
		if p == nil {
			continue
		}
		err = qtx.CreatePayout(ctx, sqlc.CreatePayoutParams{
			ID:          uuid.New().String(),
			PoolID:      poolID,
			Position:    int32(p.Position),
			AmountCents: int32(p.AmountCents),
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("creating payout for pool %s: %w", poolID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to replace payouts for pool %s: %w", poolID, err)
	}
	return nil
}
