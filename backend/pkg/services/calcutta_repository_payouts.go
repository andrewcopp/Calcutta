package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func (r *CalcuttaRepository) GetPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error) {
	query := `
		SELECT id, calcutta_id, position, amount_cents, created_at, updated_at, deleted_at
		FROM calcutta_payouts
		WHERE calcutta_id = $1 AND deleted_at IS NULL
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payouts := make([]*models.CalcuttaPayout, 0)
	for rows.Next() {
		p := &models.CalcuttaPayout{}
		var createdAt, updatedAt time.Time
		var deletedAt sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.CalcuttaID,
			&p.Position,
			&p.AmountCents,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}
		p.Created = createdAt
		p.Updated = updatedAt
		if deletedAt.Valid {
			p.Deleted = &deletedAt.Time
		}
		payouts = append(payouts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payouts, nil
}
