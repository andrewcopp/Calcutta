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

type CalcuttaRepository struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewCalcuttaRepository(pool *pgxpool.Pool) *CalcuttaRepository {
	return &CalcuttaRepository{pool: pool, q: sqlc.New(pool)}
}

func (r *CalcuttaRepository) GetAll(ctx context.Context) ([]*models.Calcutta, error) {
	rows, err := r.q.ListCalcuttas(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing calcuttas: %w", err)
	}

	out := make([]*models.Calcutta, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Calcutta{
			ID:           row.ID,
			TournamentID: row.TournamentID,
			OwnerID:      row.OwnerID,
			CreatedBy:    row.CreatedBy,
			Name:         row.Name,
			MinTeams:     int(row.MinTeams),
			MaxTeams:     int(row.MaxTeams),
			MaxBidPoints: int(row.MaxBidPoints),
			BudgetPoints: int(row.BudgetPoints),
			Visibility:   row.Visibility,
			CreatedAt:    row.CreatedAt.Time,
			UpdatedAt:    row.UpdatedAt.Time,
			DeletedAt:    nil,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Calcutta, error) {
	rows, err := r.q.ListCalcuttasByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing calcuttas by user %s: %w", userID, err)
	}

	out := make([]*models.Calcutta, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Calcutta{
			ID:           row.ID,
			TournamentID: row.TournamentID,
			OwnerID:      row.OwnerID,
			CreatedBy:    row.CreatedBy,
			Name:         row.Name,
			MinTeams:     int(row.MinTeams),
			MaxTeams:     int(row.MaxTeams),
			MaxBidPoints: int(row.MaxBidPoints),
			BudgetPoints: int(row.BudgetPoints),
			Visibility:   row.Visibility,
			CreatedAt:    row.CreatedAt.Time,
			UpdatedAt:    row.UpdatedAt.Time,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetDistinctUserIDsByCalcutta(ctx context.Context, calcuttaID string) ([]string, error) {
	uuids, err := r.q.ListDistinctUserIDsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, fmt.Errorf("listing distinct user IDs for calcutta %s: %w", calcuttaID, err)
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

func (r *CalcuttaRepository) GetByID(ctx context.Context, id string) (*models.Calcutta, error) {
	row, err := r.q.GetCalcuttaByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "calcutta", ID: id}
		}
		return nil, fmt.Errorf("getting calcutta %s: %w", id, err)
	}
	return &models.Calcutta{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		OwnerID:      row.OwnerID,
		CreatedBy:    row.CreatedBy,
		Name:         row.Name,
		MinTeams:     int(row.MinTeams),
		MaxTeams:     int(row.MaxTeams),
		MaxBidPoints: int(row.MaxBidPoints),
		BudgetPoints: int(row.BudgetPoints),
		Visibility:   row.Visibility,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		DeletedAt:    nil,
	}, nil
}

func (r *CalcuttaRepository) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	rows, err := r.q.GetCalcuttasByTournament(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("listing calcuttas by tournament %s: %w", tournamentID, err)
	}

	out := make([]*models.Calcutta, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Calcutta{
			ID:           row.ID,
			TournamentID: row.TournamentID,
			OwnerID:      row.OwnerID,
			CreatedBy:    row.CreatedBy,
			Name:         row.Name,
			MinTeams:     int(row.MinTeams),
			MaxTeams:     int(row.MaxTeams),
			MaxBidPoints: int(row.MaxBidPoints),
			BudgetPoints: int(row.BudgetPoints),
			Visibility:   row.Visibility,
			CreatedAt:    row.CreatedAt.Time,
			UpdatedAt:    row.UpdatedAt.Time,
			DeletedAt:    TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) Create(ctx context.Context, calcutta *models.Calcutta) error {
	now := time.Now()
	calcutta.ID = uuid.New().String()
	calcutta.CreatedAt = now
	calcutta.UpdatedAt = now

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to create calcutta: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	if calcutta.Visibility == "" {
		calcutta.Visibility = "private"
	}
	params := sqlc.CreateCalcuttaParams{
		ID:           calcutta.ID,
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		CreatedBy:    calcutta.CreatedBy,
		Name:         calcutta.Name,
		MinTeams:     int32(calcutta.MinTeams),
		MaxTeams:     int32(calcutta.MaxTeams),
		MaxBidPoints:       int32(calcutta.MaxBidPoints),
		BudgetPoints: int32(calcutta.BudgetPoints),
		Visibility:   calcutta.Visibility,
		CreatedAt:    pgtype.Timestamptz{Time: calcutta.CreatedAt, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: calcutta.UpdatedAt, Valid: true},
	}
	if err = qtx.CreateCalcutta(ctx, params); err != nil {
		return fmt.Errorf("creating calcutta: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to create calcutta: %w", err)
	}
	return nil
}

func (r *CalcuttaRepository) Update(ctx context.Context, calcutta *models.Calcutta) error {
	calcutta.UpdatedAt = time.Now()

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to update calcutta %s: %w", calcutta.ID, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.UpdateCalcuttaParams{
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		Name:         calcutta.Name,
		MinTeams:     int32(calcutta.MinTeams),
		MaxTeams:     int32(calcutta.MaxTeams),
		MaxBidPoints:       int32(calcutta.MaxBidPoints),
		BudgetPoints: int32(calcutta.BudgetPoints),
		Visibility:   calcutta.Visibility,
		UpdatedAt:    pgtype.Timestamptz{Time: calcutta.UpdatedAt, Valid: true},
		ID:           calcutta.ID,
	}
	affected, err := qtx.UpdateCalcutta(ctx, params)
	if err != nil {
		return fmt.Errorf("updating calcutta %s: %w", calcutta.ID, err)
	}
	if affected == 0 {
		err = &apperrors.NotFoundError{Resource: "calcutta", ID: calcutta.ID}
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to update calcutta %s: %w", calcutta.ID, err)
	}
	return nil
}

func (r *CalcuttaRepository) GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error) {
	rows, err := r.q.ListCalcuttaRounds(ctx, calcuttaID)
	if err != nil {
		return nil, fmt.Errorf("listing calcutta rounds for calcutta %s: %w", calcuttaID, err)
	}

	out := make([]*models.CalcuttaRound, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaRound{
			ID:         row.ID,
			CalcuttaID: row.CalcuttaID,
			Round:      int(row.Round),
			Points:     int(row.Points),
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
			DeletedAt:  nil,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) CreateRound(ctx context.Context, round *models.CalcuttaRound) error {
	now := time.Now()
	round.ID = uuid.New().String()
	round.CreatedAt = now
	round.UpdatedAt = now

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to create calcutta round: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	params := sqlc.CreateCalcuttaRoundParams{
		ID:            round.ID,
		CalcuttaID:    round.CalcuttaID,
		WinIndex:      int32(round.Round),
		PointsAwarded: int32(round.Points),
		CreatedAt:     pgtype.Timestamptz{Time: round.CreatedAt, Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: round.UpdatedAt, Valid: true},
	}
	if err = qtx.CreateCalcuttaRound(ctx, params); err != nil {
		return fmt.Errorf("creating calcutta round: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to create calcutta round: %w", err)
	}
	return nil
}
