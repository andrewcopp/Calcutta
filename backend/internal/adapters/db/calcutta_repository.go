package db

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/apperrors"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
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
		return nil, err
	}

	out := make([]*models.Calcutta, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Calcutta{
			ID:           row.ID,
			TournamentID: row.TournamentID,
			OwnerID:      row.OwnerID,
			Name:         row.Name,
			MinTeams:     int(row.MinTeams),
			MaxTeams:     int(row.MaxTeams),
			MaxBid:       int(row.MaxBid),
			Created:      row.CreatedAt.Time,
			Updated:      row.UpdatedAt.Time,
			Deleted:      nil,
		})
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

func (r *CalcuttaRepository) GetByID(ctx context.Context, id string) (*models.Calcutta, error) {
	row, err := r.q.GetCalcuttaByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "calcutta", ID: id}
		}
		return nil, err
	}
	return &models.Calcutta{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		OwnerID:      row.OwnerID,
		Name:         row.Name,
		MinTeams:     int(row.MinTeams),
		MaxTeams:     int(row.MaxTeams),
		MaxBid:       int(row.MaxBid),
		Created:      row.CreatedAt.Time,
		Updated:      row.UpdatedAt.Time,
		Deleted:      nil,
	}, nil
}

func (r *CalcuttaRepository) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	rows, err := r.q.GetCalcuttasByTournament(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.Calcutta, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Calcutta{
			ID:           row.ID,
			TournamentID: row.TournamentID,
			OwnerID:      row.OwnerID,
			Name:         row.Name,
			MinTeams:     int(row.MinTeams),
			MaxTeams:     int(row.MaxTeams),
			MaxBid:       int(row.MaxBid),
			Created:      row.CreatedAt.Time,
			Updated:      row.UpdatedAt.Time,
			Deleted:      timestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) Create(ctx context.Context, calcutta *models.Calcutta) error {
	now := time.Now()
	calcutta.ID = uuid.New().String()
	calcutta.Created = now
	calcutta.Updated = now

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
	params := sqlc.CreateCalcuttaParams{
		ID:           calcutta.ID,
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		Name:         calcutta.Name,
		MinTeams:     int32(calcutta.MinTeams),
		MaxTeams:     int32(calcutta.MaxTeams),
		MaxBid:       int32(calcutta.MaxBid),
		CreatedAt:    pgtype.Timestamptz{Time: calcutta.Created, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: calcutta.Updated, Valid: true},
	}
	if err = qtx.CreateCalcutta(ctx, params); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CalcuttaRepository) Update(ctx context.Context, calcutta *models.Calcutta) error {
	calcutta.Updated = time.Now()

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
	params := sqlc.UpdateCalcuttaParams{
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		Name:         calcutta.Name,
		MinTeams:     int32(calcutta.MinTeams),
		MaxTeams:     int32(calcutta.MaxTeams),
		MaxBid:       int32(calcutta.MaxBid),
		UpdatedAt:    pgtype.Timestamptz{Time: calcutta.Updated, Valid: true},
		ID:           calcutta.ID,
	}
	affected, err := qtx.UpdateCalcutta(ctx, params)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "calcutta", ID: calcutta.ID}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CalcuttaRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()

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
	params := sqlc.DeleteCalcuttaParams{
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		ID:        id,
	}
	affected, err := qtx.DeleteCalcutta(ctx, params)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "calcutta", ID: id}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *CalcuttaRepository) GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error) {
	rows, err := r.q.ListCalcuttaRounds(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaRound, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaRound{
			ID:         row.ID,
			CalcuttaID: row.CalcuttaID,
			Round:      int(row.Round),
			Points:     int(row.Points),
			Created:    row.CreatedAt.Time,
			Updated:    row.UpdatedAt.Time,
			Deleted:    nil,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) CreateRound(ctx context.Context, round *models.CalcuttaRound) error {
	now := time.Now()
	round.ID = uuid.New().String()
	round.Created = now
	round.Updated = now

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
	params := sqlc.CreateCalcuttaRoundParams{
		ID:            round.ID,
		CalcuttaID:    round.CalcuttaID,
		WinIndex:      int32(round.Round),
		PointsAwarded: int32(round.Points),
		CreatedAt:     pgtype.Timestamptz{Time: round.Created, Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: round.Updated, Valid: true},
	}
	if err = qtx.CreateCalcuttaRound(ctx, params); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

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
			Deleted:     timestamptzToPtrTime(row.DeletedAt),
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
		Deleted:    timestamptzToPtrTime(row.DeletedAt),
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
			Deleted: timestamptzToPtrTime(row.DeletedAt),
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
			Deleted:      timestamptzToPtrTime(row.TeamDeletedAt),
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		team.Team = tt

		out = append(out, team)
	}
	return out, nil
}

func (r *CalcuttaRepository) GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error) {
	row, err := r.q.GetPortfolioByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "portfolio", ID: id}
		}
		return nil, err
	}

	return &models.CalcuttaPortfolio{
		ID:            row.ID,
		EntryID:       row.EntryID,
		MaximumPoints: row.MaximumPoints,
		Created:       row.CreatedAt.Time,
		Updated:       row.UpdatedAt.Time,
		Deleted:       timestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *CalcuttaRepository) GetPortfolioTeams(ctx context.Context, portfolioID string) ([]*models.CalcuttaPortfolioTeam, error) {
	rows, err := r.q.ListPortfolioTeamsByPortfolioID(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaPortfolioTeam, 0, len(rows))
	for _, row := range rows {
		pt := &models.CalcuttaPortfolioTeam{
			ID:                  row.ID,
			PortfolioID:         row.PortfolioID,
			TeamID:              row.TeamID,
			OwnershipPercentage: row.OwnershipPercentage,
			ActualPoints:        row.ActualPoints,
			ExpectedPoints:      row.ExpectedPoints,
			PredictedPoints:     row.PredictedPoints,
			Created:             row.CreatedAt.Time,
			Updated:             row.UpdatedAt.Time,
			Deleted:             timestamptzToPtrTime(row.DeletedAt),
		}

		tt := &models.TournamentTeam{
			ID:           row.TournamentTeamID,
			SchoolID:     row.SchoolID,
			TournamentID: row.TournamentID,
			Seed:         int(row.Seed),
			Region:       row.Region,
			Byes:         int(row.Byes),
			Wins:         int(row.Wins),
			Eliminated:   row.Eliminated,
			Created:      row.TeamCreatedAt.Time,
			Updated:      row.TeamUpdatedAt.Time,
		}
		if row.SchoolName != nil {
			tt.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
		}
		pt.Team = tt

		out = append(out, pt)
	}
	return out, nil
}

func (r *CalcuttaRepository) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	rows, err := r.q.ListPortfoliosByEntryID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaPortfolio, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaPortfolio{
			ID:      row.ID,
			EntryID: row.EntryID,
			Created: row.CreatedAt.Time,
			Updated: row.UpdatedAt.Time,
			Deleted: timestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetPortfolios(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	rows, err := r.q.ListPortfolios(ctx, entryID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.CalcuttaPortfolio, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.CalcuttaPortfolio{
			ID:            row.ID,
			EntryID:       row.EntryID,
			MaximumPoints: row.MaximumPoints,
			Created:       row.CreatedAt.Time,
			Updated:       row.UpdatedAt.Time,
			Deleted:       timestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	row, err := r.q.GetTournamentTeamByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &apperrors.NotFoundError{Resource: "tournament team", ID: id}
		}
		return nil, err
	}

	team := &models.TournamentTeam{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		SchoolID:     row.SchoolID,
		Seed:         int(row.Seed),
		Region:       row.Region,
		Byes:         int(row.Byes),
		Wins:         int(row.Wins),
		Eliminated:   row.Eliminated,
		Created:      row.CreatedAt.Time,
		Updated:      row.UpdatedAt.Time,
	}
	if row.NetRtg != nil || row.ORtg != nil || row.DRtg != nil || row.AdjT != nil {
		team.KenPom = &models.KenPomStats{NetRtg: row.NetRtg, ORtg: row.ORtg, DRtg: row.DRtg, AdjT: row.AdjT}
	}
	if row.SchoolName != nil {
		team.School = &models.School{ID: row.SchoolID, Name: *row.SchoolName}
	}
	return team, nil
}

func timestamptzToPtrTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func uuidToPtrString(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	s := id.String()
	return &s
}

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	// pgtype.Numeric.Scan does not reliably accept float64 across pgx versions.
	// Format as a decimal string to ensure consistent behavior.
	s := strconv.FormatFloat(v, 'f', -1, 64)
	err := n.Scan(s)
	return n, err
}
