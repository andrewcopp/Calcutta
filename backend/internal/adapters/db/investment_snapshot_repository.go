package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.InvestmentSnapshotWriter = (*InvestmentSnapshotRepository)(nil)

type InvestmentSnapshotRepository struct {
	q *sqlc.Queries
}

func NewInvestmentSnapshotRepository(pool *pgxpool.Pool) *InvestmentSnapshotRepository {
	return &InvestmentSnapshotRepository{q: sqlc.New(pool)}
}

func (r *InvestmentSnapshotRepository) CreateInvestmentSnapshot(ctx context.Context, snapshot *models.InvestmentSnapshot) error {
	investmentsJSON, err := json.Marshal(snapshot.Investments)
	if err != nil {
		return fmt.Errorf("marshalling investments: %w", err)
	}

	if err := r.q.CreateInvestmentSnapshot(ctx, sqlc.CreateInvestmentSnapshotParams{
		PortfolioID: snapshot.PortfolioID,
		ChangedBy:   snapshot.ChangedBy,
		Reason:      snapshot.Reason,
		Investments: investmentsJSON,
	}); err != nil {
		return fmt.Errorf("creating investment snapshot: %w", err)
	}
	return nil
}
