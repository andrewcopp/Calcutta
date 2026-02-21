package db

import (
	"context"
	"errors"
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
	entry.Created = now
	entry.Updated = now

	var userID pgtype.UUID
	if entry.UserID != nil {
		parsed, err := uuid.Parse(*entry.UserID)
		if err != nil {
			return err
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
		return err
	}
	return nil
}

func (r *CalcuttaRepository) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	rows, err := r.q.ListEntriesByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaEntry{
			ID:          row.ID,
			Name:        row.Name,
			UserID:      uuidToPtrString(row.UserID),
			CalcuttaID:  row.CalcuttaID,
			TotalPoints: row.TotalPoints,
			Created:     row.CreatedAt.Time,
			Updated:     row.UpdatedAt.Time,
			Deleted:     TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	row, err := r.q.GetEntryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "entry", ID: id}
		}
		return nil, err
	}

	return &models.CalcuttaEntry{
		ID:         row.ID,
		Name:       row.Name,
		UserID:     uuidToPtrString(row.UserID),
		CalcuttaID: row.CalcuttaID,
		Created:    row.CreatedAt.Time,
		Updated:    row.UpdatedAt.Time,
		Deleted:    TimestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *CalcuttaRepository) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	rows, err := r.q.ListEntryTeamsByEntryID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaEntryTeam, 0, len(rows))
	for _, row := range rows {
		team := &models.CalcuttaEntryTeam{
			ID:      row.ID,
			EntryID: row.EntryID,
			TeamID:  row.TeamID,
			Bid:     int(row.Bid),
			Created: row.CreatedAt.Time,
			Updated: row.UpdatedAt.Time,
			Deleted: TimestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			Created:      row.TeamCreatedAt.Time,
			Updated:      row.TeamUpdatedAt.Time,
			Deleted:      TimestamptzToPtrTime(row.TeamDeletedAt),
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		team.Team = tt

		out = append(out, team)
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
		return err
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
		return err
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
			BidPoints: int32(t.Bid),
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}
		if err = qtx.CreateEntryTeam(ctx, params); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
