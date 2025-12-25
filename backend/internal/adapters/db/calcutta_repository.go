package db

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
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
			Created:      row.CreatedAt.Time,
			Updated:      row.UpdatedAt.Time,
			Deleted:      nil,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetByID(ctx context.Context, id string) (*models.Calcutta, error) {
	row, err := r.q.GetCalcuttaByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &services.NotFoundError{Resource: "calcutta", ID: id}
		}
		return nil, err
	}
	return &models.Calcutta{
		ID:           row.ID,
		TournamentID: row.TournamentID,
		OwnerID:      row.OwnerID,
		Name:         row.Name,
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

	return r.q.CreateCalcutta(ctx, sqlc.CreateCalcuttaParams{
		ID:           calcutta.ID,
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		Name:         calcutta.Name,
		CreatedAt:    pgtype.Timestamptz{Time: calcutta.Created, Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: calcutta.Updated, Valid: true},
	})
}

func (r *CalcuttaRepository) Update(ctx context.Context, calcutta *models.Calcutta) error {
	calcutta.Updated = time.Now()

	affected, err := r.q.UpdateCalcutta(ctx, sqlc.UpdateCalcuttaParams{
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		Name:         calcutta.Name,
		UpdatedAt:    pgtype.Timestamptz{Time: calcutta.Updated, Valid: true},
		ID:           calcutta.ID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return &services.NotFoundError{Resource: "calcutta", ID: calcutta.ID}
	}
	return nil
}

func (r *CalcuttaRepository) Delete(ctx context.Context, id string) error {
	now := time.Now()
	affected, err := r.q.DeleteCalcutta(ctx, sqlc.DeleteCalcuttaParams{
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		ID:        id,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return &services.NotFoundError{Resource: "calcutta", ID: id}
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

	return r.q.CreateCalcuttaRound(ctx, sqlc.CreateCalcuttaRoundParams{
		ID:         round.ID,
		CalcuttaID: round.CalcuttaID,
		Round:      int32(round.Round),
		Points:     int32(round.Points),
		CreatedAt:  pgtype.Timestamptz{Time: round.Created, Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: round.Updated, Valid: true},
	})
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
			return nil, &services.NotFoundError{Resource: "entry", ID: id}
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
			return nil, &services.NotFoundError{Resource: "portfolio", ID: id}
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

func (r *CalcuttaRepository) CreatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error {
	now := time.Now()
	portfolio.ID = uuid.New().String()
	portfolio.Created = now
	portfolio.Updated = now

	maxPts, err := numericFromFloat64(portfolio.MaximumPoints)
	if err != nil {
		return err
	}

	return r.q.CreatePortfolio(ctx, sqlc.CreatePortfolioParams{
		ID:            portfolio.ID,
		EntryID:       portfolio.EntryID,
		MaximumPoints: maxPts,
		CreatedAt:     pgtype.Timestamptz{Time: portfolio.Created, Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: portfolio.Updated, Valid: true},
	})
}

func (r *CalcuttaRepository) UpdatePortfolio(ctx context.Context, portfolio *models.CalcuttaPortfolio) error {
	maxPts, err := numericFromFloat64(portfolio.MaximumPoints)
	if err != nil {
		return err
	}

	affected, err := r.q.UpdatePortfolio(ctx, sqlc.UpdatePortfolioParams{
		MaximumPoints: maxPts,
		UpdatedAt:     pgtype.Timestamptz{Time: portfolio.Updated, Valid: true},
		ID:            portfolio.ID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return &services.NotFoundError{Resource: "portfolio", ID: portfolio.ID}
	}
	return nil
}

func (r *CalcuttaRepository) CreatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	now := time.Now()
	team.ID = uuid.New().String()
	team.Created = now
	team.Updated = now

	op, err := numericFromFloat64(team.OwnershipPercentage)
	if err != nil {
		return err
	}
	ap, err := numericFromFloat64(team.ActualPoints)
	if err != nil {
		return err
	}
	ep, err := numericFromFloat64(team.ExpectedPoints)
	if err != nil {
		return err
	}
	pp, err := numericFromFloat64(team.PredictedPoints)
	if err != nil {
		return err
	}

	return r.q.CreatePortfolioTeam(ctx, sqlc.CreatePortfolioTeamParams{
		ID:                  team.ID,
		PortfolioID:         team.PortfolioID,
		TeamID:              team.TeamID,
		OwnershipPercentage: op,
		ActualPoints:        ap,
		ExpectedPoints:      ep,
		PredictedPoints:     pp,
		CreatedAt:           pgtype.Timestamptz{Time: team.Created, Valid: true},
		UpdatedAt:           pgtype.Timestamptz{Time: team.Updated, Valid: true},
	})
}

func (r *CalcuttaRepository) UpdatePortfolioTeam(ctx context.Context, team *models.CalcuttaPortfolioTeam) error {
	op, err := numericFromFloat64(team.OwnershipPercentage)
	if err != nil {
		return err
	}
	ap, err := numericFromFloat64(team.ActualPoints)
	if err != nil {
		return err
	}
	ep, err := numericFromFloat64(team.ExpectedPoints)
	if err != nil {
		return err
	}
	pp, err := numericFromFloat64(team.PredictedPoints)
	if err != nil {
		return err
	}

	affected, err := r.q.UpdatePortfolioTeam(ctx, sqlc.UpdatePortfolioTeamParams{
		OwnershipPercentage: op,
		ActualPoints:        ap,
		ExpectedPoints:      ep,
		PredictedPoints:     pp,
		UpdatedAt:           pgtype.Timestamptz{Time: team.Updated, Valid: true},
		ID:                  team.ID,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return &services.NotFoundError{Resource: "portfolio team", ID: team.ID}
	}
	return nil
}

func (r *CalcuttaRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	row, err := r.q.GetTournamentTeamByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &services.NotFoundError{Resource: "tournament team", ID: id}
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
