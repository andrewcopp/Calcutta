package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewPoolRepository(pool *pgxpool.Pool) *PoolRepository {
	return &PoolRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *PoolRepository) GetAll(ctx context.Context) ([]*models.Pool, error) {
	rows, err := r.q.ListPools(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing pools: %w", err)
	}

	out := make([]*models.Pool, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Pool{
			ID:                   row.ID,
			TournamentID:         row.TournamentID,
			OwnerID:              row.OwnerID,
			CreatedBy:            row.CreatedBy,
			Name:                 row.Name,
			MinTeams:             int(row.MinTeams),
			MaxTeams:             int(row.MaxTeams),
			MaxInvestmentCredits: int(row.MaxInvestmentCredits),
			BudgetCredits:        int(row.BudgetCredits),
			Visibility:           row.Visibility,
			CreatedAt:            row.CreatedAt.Time,
			UpdatedAt:            row.UpdatedAt.Time,
			DeletedAt:            nil,
		})
	}
	return out, nil
}

func (r *PoolRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Pool, error) {
	rows, err := r.q.ListPoolsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing pools by user %s: %w", userID, err)
	}

	out := make([]*models.Pool, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Pool{
			ID:                   row.ID,
			TournamentID:         row.TournamentID,
			OwnerID:              row.OwnerID,
			CreatedBy:            row.CreatedBy,
			Name:                 row.Name,
			MinTeams:             int(row.MinTeams),
			MaxTeams:             int(row.MaxTeams),
			MaxInvestmentCredits: int(row.MaxInvestmentCredits),
			BudgetCredits:        int(row.BudgetCredits),
			Visibility:           row.Visibility,
			CreatedAt:            row.CreatedAt.Time,
			UpdatedAt:            row.UpdatedAt.Time,
		})
	}
	return out, nil
}

func (r *PoolRepository) GetDistinctUserIDsByPool(ctx context.Context, poolID string) ([]string, error) {
	uuids, err := r.q.ListDistinctUserIDsByPoolID(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("listing distinct user IDs for pool %s: %w", poolID, err)
	}
	out := make([]string, 0, len(uuids))
	for _, u := range uuids {
		s := uuidToPtrString(u)
		if s != nil {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (r *PoolRepository) GetByID(ctx context.Context, id string) (*models.Pool, error) {
	row, err := r.q.GetPoolByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "pool", ID: id}
		}
		return nil, fmt.Errorf("getting pool %s: %w", id, err)
	}
	return &models.Pool{
		ID:                   row.ID,
		TournamentID:         row.TournamentID,
		OwnerID:              row.OwnerID,
		CreatedBy:            row.CreatedBy,
		Name:                 row.Name,
		MinTeams:             int(row.MinTeams),
		MaxTeams:             int(row.MaxTeams),
		MaxInvestmentCredits: int(row.MaxInvestmentCredits),
		BudgetCredits:        int(row.BudgetCredits),
		Visibility:           row.Visibility,
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
		DeletedAt:            nil,
	}, nil
}

func (r *PoolRepository) GetPoolsByTournament(ctx context.Context, tournamentID string) ([]*models.Pool, error) {
	rows, err := r.q.GetPoolsByTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("listing pools by tournament %s: %w", tournamentID, err)
	}

	out := make([]*models.Pool, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Pool{
			ID:                   row.ID,
			TournamentID:         row.TournamentID,
			OwnerID:              row.OwnerID,
			CreatedBy:            row.CreatedBy,
			Name:                 row.Name,
			MinTeams:             int(row.MinTeams),
			MaxTeams:             int(row.MaxTeams),
			MaxInvestmentCredits: int(row.MaxInvestmentCredits),
			BudgetCredits:        int(row.BudgetCredits),
			Visibility:           row.Visibility,
			CreatedAt:            row.CreatedAt.Time,
			UpdatedAt:            row.UpdatedAt.Time,
			DeletedAt:            TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *PoolRepository) Create(ctx context.Context, pool *models.Pool) error {
	now := time.Now()
	pool.ID = uuid.New().String()
	pool.CreatedAt = now
	pool.UpdatedAt = now

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to create pool: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	if pool.Visibility == "" {
		pool.Visibility = "private"
	}
	params := sqlc.CreatePoolParams{
		ID:                   pool.ID,
		TournamentID:         pool.TournamentID,
		OwnerID:              pool.OwnerID,
		CreatedBy:            pool.CreatedBy,
		Name:                 pool.Name,
		MinTeams:             int32(pool.MinTeams),
		MaxTeams:             int32(pool.MaxTeams),
		MaxInvestmentCredits: int32(pool.MaxInvestmentCredits),
		BudgetCredits:        int32(pool.BudgetCredits),
		Visibility:           pool.Visibility,
		CreatedAt:            pgtype.Timestamptz{Time: pool.CreatedAt, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: pool.UpdatedAt, Valid: true},
	}
	if err = qtx.CreatePool(ctx, params); err != nil {
		return fmt.Errorf("creating pool: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to create pool: %w", err)
	}
	return nil
}

func (r *PoolRepository) Update(ctx context.Context, pool *models.Pool) error {
	pool.UpdatedAt = time.Now()

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to update pool %s: %w", pool.ID, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.UpdatePoolParams{
		TournamentID:         pool.TournamentID,
		OwnerID:              pool.OwnerID,
		Name:                 pool.Name,
		MinTeams:             int32(pool.MinTeams),
		MaxTeams:             int32(pool.MaxTeams),
		MaxInvestmentCredits: int32(pool.MaxInvestmentCredits),
		BudgetCredits:        int32(pool.BudgetCredits),
		Visibility:           pool.Visibility,
		UpdatedAt:            pgtype.Timestamptz{Time: pool.UpdatedAt, Valid: true},
		ID:                   pool.ID,
	}
	affected, err := qtx.UpdatePool(ctx, params)
	if err != nil {
		return fmt.Errorf("updating pool %s: %w", pool.ID, err)
	}
	if affected == 0 {
		err = &apperrors.NotFoundError{Resource: "pool", ID: pool.ID}
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to update pool %s: %w", pool.ID, err)
	}
	return nil
}

func (r *PoolRepository) GetScoringRules(ctx context.Context, poolID string) ([]*models.ScoringRule, error) {
	rows, err := r.q.ListScoringRules(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("listing scoring rules for pool %s: %w", poolID, err)
	}

	out := make([]*models.ScoringRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.ScoringRule{
			ID:            row.ID,
			PoolID:        row.PoolID,
			WinIndex:      int(row.Round),
			PointsAwarded: int(row.Points),
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			DeletedAt:     nil,
		})
	}
	return out, nil
}

func (r *PoolRepository) CreateScoringRule(ctx context.Context, rule *models.ScoringRule) error {
	now := time.Now()
	rule.ID = uuid.New().String()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to create scoring rule: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.CreateScoringRuleParams{
		ID:            rule.ID,
		PoolID:        rule.PoolID,
		WinIndex:      int32(rule.WinIndex),
		PointsAwarded: int32(rule.PointsAwarded),
		CreatedAt:     pgtype.Timestamptz{Time: rule.CreatedAt, Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: rule.UpdatedAt, Valid: true},
	}
	if err = qtx.CreateScoringRule(ctx, params); err != nil {
		return fmt.Errorf("creating scoring rule: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to create scoring rule: %w", err)
	}
	return nil
}
