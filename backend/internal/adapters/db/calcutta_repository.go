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
			ID:              row.ID,
			TournamentID:    row.TournamentID,
			OwnerID:         row.OwnerID,
			Name:            row.Name,
			MinTeams:        int(row.MinTeams),
			MaxTeams:        int(row.MaxTeams),
			MaxBid:          int(row.MaxBid),
			BiddingOpen:     row.BiddingOpen,
			BiddingLockedAt: TimestamptzToPtrTime(row.BiddingLockedAt),
			Visibility:      row.Visibility,
			Created:         row.CreatedAt.Time,
			Updated:         row.UpdatedAt.Time,
			Deleted:         nil,
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

func (r *CalcuttaRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Calcutta, error) {
	rows, err := r.q.ListCalcuttasByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*models.Calcutta, 0, len(rows))
	for _, row := range rows {
		out = append(out, &models.Calcutta{
			ID:              row.ID,
			TournamentID:    row.TournamentID,
			OwnerID:         row.OwnerID,
			Name:            row.Name,
			MinTeams:        int(row.MinTeams),
			MaxTeams:        int(row.MaxTeams),
			MaxBid:          int(row.MaxBid),
			BiddingOpen:     row.BiddingOpen,
			BiddingLockedAt: TimestamptzToPtrTime(row.BiddingLockedAt),
			Visibility:      row.Visibility,
			Created:         row.CreatedAt.Time,
			Updated:         row.UpdatedAt.Time,
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetDistinctUserIDsByCalcutta(ctx context.Context, calcuttaID string) ([]string, error) {
	uuids, err := r.q.ListDistinctUserIDsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return &models.Calcutta{
		ID:              row.ID,
		TournamentID:    row.TournamentID,
		OwnerID:         row.OwnerID,
		Name:            row.Name,
		MinTeams:        int(row.MinTeams),
		MaxTeams:        int(row.MaxTeams),
		MaxBid:          int(row.MaxBid),
		BiddingOpen:     row.BiddingOpen,
		BiddingLockedAt: TimestamptzToPtrTime(row.BiddingLockedAt),
		Visibility:      row.Visibility,
		Created:         row.CreatedAt.Time,
		Updated:         row.UpdatedAt.Time,
		Deleted:         nil,
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
			ID:              row.ID,
			TournamentID:    row.TournamentID,
			OwnerID:         row.OwnerID,
			Name:            row.Name,
			MinTeams:        int(row.MinTeams),
			MaxTeams:        int(row.MaxTeams),
			MaxBid:          int(row.MaxBid),
			BiddingOpen:     row.BiddingOpen,
			BiddingLockedAt: TimestamptzToPtrTime(row.BiddingLockedAt),
			Visibility:      row.Visibility,
			Created:         row.CreatedAt.Time,
			Updated:         row.UpdatedAt.Time,
			Deleted:         TimestamptzToPtrTime(row.DeletedAt),
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
	if calcutta.Visibility == "" {
		calcutta.Visibility = "private"
	}
	params := sqlc.CreateCalcuttaParams{
		ID:           calcutta.ID,
		TournamentID: calcutta.TournamentID,
		OwnerID:      calcutta.OwnerID,
		Name:         calcutta.Name,
		MinTeams:     int32(calcutta.MinTeams),
		MaxTeams:     int32(calcutta.MaxTeams),
		MaxBid:       int32(calcutta.MaxBid),
		Visibility:   calcutta.Visibility,
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
	var biddingLockedAt pgtype.Timestamptz
	if calcutta.BiddingLockedAt != nil {
		biddingLockedAt = pgtype.Timestamptz{Time: *calcutta.BiddingLockedAt, Valid: true}
	}
	params := sqlc.UpdateCalcuttaParams{
		TournamentID:    calcutta.TournamentID,
		OwnerID:         calcutta.OwnerID,
		Name:            calcutta.Name,
		MinTeams:        int32(calcutta.MinTeams),
		MaxTeams:        int32(calcutta.MaxTeams),
		MaxBid:          int32(calcutta.MaxBid),
		BiddingOpen:     calcutta.BiddingOpen,
		BiddingLockedAt: biddingLockedAt,
		Visibility:      calcutta.Visibility,
		UpdatedAt:       pgtype.Timestamptz{Time: calcutta.Updated, Valid: true},
		ID:              calcutta.ID,
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

func (r *CalcuttaRepository) ReplacePayouts(ctx context.Context, calcuttaID string, payouts []*models.CalcuttaPayout) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
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
		return err
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
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

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

	if entry.Status == "" {
		entry.Status = "draft"
	}
	params := sqlc.CreateEntryParams{
		ID:         entry.ID,
		Name:       entry.Name,
		UserID:     userID,
		CalcuttaID: entry.CalcuttaID,
		Status:     entry.Status,
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
			Status:      row.Status,
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
		Status:     row.Status,
		Created:    row.CreatedAt.Time,
		Updated:    row.UpdatedAt.Time,
		Deleted:    TimestamptzToPtrTime(row.DeletedAt),
	}, nil
}

func (r *CalcuttaRepository) UpdateEntryStatus(ctx context.Context, id string, status string) error {
	affected, err := r.q.UpdateEntryStatus(ctx, sqlc.UpdateEntryStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return &apperrors.NotFoundError{Resource: "entry", ID: id}
	}
	return nil
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
		Deleted:       TimestamptzToPtrTime(row.DeletedAt),
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
			Created:             row.CreatedAt.Time,
			Updated:             row.UpdatedAt.Time,
			Deleted:             TimestamptzToPtrTime(row.DeletedAt),
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
			Deleted: TimestamptzToPtrTime(row.DeletedAt),
		})
	}
	return out, nil
}

func (r *CalcuttaRepository) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	row, err := r.q.GetTeamByID(ctx, id)
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
