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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *CalcuttaRepository) CreateEntry(ctx context.Context, entry *models.CalcuttaEntry) error {
	entry.ID = uuid.New().String()
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	var userID pgtype.UUID
	if entry.UserID != nil {
		parsed, err := uuid.Parse(*entry.UserID)
		if err != nil {
			return fmt.Errorf("parsing user ID for entry: %w", err)
		}
		userID = pgtype.UUID{Bytes: parsed, Valid: true}
	}

	params := sqlc.CreateEntryParams{
		ID:         entry.ID,
		Name:       entry.Name,
		UserID:     userID,
		CalcuttaID: entry.CalcuttaID,
	}
	if err := r.q.CreateEntry(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &apperrors.AlreadyExistsError{Resource: "entry", Field: "user_id"}
		}
		return fmt.Errorf("creating entry: %w", err)
	}
	return nil
}

func (r *CalcuttaRepository) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, map[string]float64, error) {
	rows, err := r.q.ListEntriesByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, nil, fmt.Errorf("listing entries for calcutta %s: %w", calcuttaID, err)
	}

	entries := make([]*models.CalcuttaEntry, 0, len(rows))
	pointsByEntry := make(map[string]float64, len(rows))
	for _, row := range rows {
		entries = append(entries, &models.CalcuttaEntry{
			ID:         row.ID,
			Name:       row.Name,
			UserID:     uuidToPtrString(row.UserID),
			CalcuttaID: row.CalcuttaID,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
			DeletedAt:  TimestamptzToPtrTime(row.DeletedAt),
		})
		pointsByEntry[row.ID] = row.TotalPoints
	}
	return entries, pointsByEntry, nil
}

func (r *CalcuttaRepository) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	row, err := r.q.GetEntryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "entry", ID: id}
		}
		return nil, fmt.Errorf("getting entry %s: %w", id, err)
	}

	return &models.CalcuttaEntry{
		ID:         row.ID,
		Name:       row.Name,
		UserID:     uuidToPtrString(row.UserID),
		CalcuttaID: row.CalcuttaID,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
		DeletedAt:  TimestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *CalcuttaRepository) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	rows, err := r.q.ListEntryTeamsByEntryID(ctx, entryID)
	if err != nil {
		return nil, fmt.Errorf("listing entry teams for entry %s: %w", entryID, err)
	}

	out := make([]*models.CalcuttaEntryTeam, 0, len(rows))
	for _, row := range rows {
		team := &models.CalcuttaEntryTeam{
			ID:        row.ID,
			EntryID:   row.EntryID,
			TeamID:    row.TeamID,
			BidPoints: int(row.BidPoints),
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
			DeletedAt:    TimestamptzToPtrTime(row.TeamDeletedAt),
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		team.Team = tt

		out = append(out, team)
	}
	return out, nil
}

func (r *CalcuttaRepository) GetEntryTeamsByEntryIDs(ctx context.Context, entryIDs []string) (map[string][]*models.CalcuttaEntryTeam, error) {
	if len(entryIDs) == 0 {
		return map[string][]*models.CalcuttaEntryTeam{}, nil
	}

	rows, err := r.q.ListEntryTeamsByEntryIDs(ctx, entryIDs)
	if err != nil {
		return nil, fmt.Errorf("listing entry teams by entry IDs: %w", err)
	}

	out := make(map[string][]*models.CalcuttaEntryTeam, len(entryIDs))
	for _, row := range rows {
		team := &models.CalcuttaEntryTeam{
			ID:        row.ID,
			EntryID:   row.EntryID,
			TeamID:    row.TeamID,
			BidPoints: int(row.BidPoints),
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
			DeletedAt: TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			CreatedAt:    row.TeamCreatedAt.Time,
			UpdatedAt:    row.TeamUpdatedAt.Time,
			DeletedAt:    TimestamptzToPtrTime(row.TeamDeletedAt),
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		team.Team = tt

		out[row.EntryID] = append(out[row.EntryID], team)
	}
	return out, nil
}

func (r *CalcuttaRepository) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	// Validate that entry exists (and that caller has access is handled at higher layers)
	if _, err := r.GetEntry(ctx, entryID); err != nil {
		return err
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("beginning transaction to replace entry teams for entry %s: %w", entryID, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	qtx := r.q.WithTx(tx)
	now := time.Now()

	if _, err = qtx.SoftDeleteEntryTeamsByEntryID(ctx, sqlc.SoftDeleteEntryTeamsByEntryIDParams{
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
		EntryID:   entryID,
	}); err != nil {
		return fmt.Errorf("soft-deleting entry teams for entry %s: %w", entryID, err)
	}

	for _, t := range teams {
		if t == nil {
			continue
		}
		id := uuid.New().String()
		params := sqlc.CreateEntryTeamParams{
			ID:        id,
			EntryID:   entryID,
			TeamID:    t.TeamID,
			BidPoints: int32(t.BidPoints),
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}
		if err = qtx.CreateEntryTeam(ctx, params); err != nil {
			return fmt.Errorf("creating entry team for entry %s: %w", entryID, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction to replace entry teams for entry %s: %w", entryID, err)
	}
	return nil
}
