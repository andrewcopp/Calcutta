package services

import (
	"context"
	"database/sql"
	"log"
)

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

type SeedAnalyticsData struct {
	Seed            int
	TotalPoints     float64
	TotalInvestment float64
	TeamCount       int
}

type RegionAnalyticsData struct {
	Region          string
	TotalPoints     float64
	TotalInvestment float64
	TeamCount       int
}

type TeamAnalyticsData struct {
	SchoolID        string
	SchoolName      string
	TotalPoints     float64
	TotalInvestment float64
	Appearances     int
	TotalSeed       int
}

type SeedVarianceData struct {
	Seed             int
	InvestmentStdDev float64
	PointsStdDev     float64
	InvestmentMean   float64
	PointsMean       float64
	TeamCount        int
}

type SeedInvestmentPointData struct {
	Seed             int
	TournamentName   string
	TournamentYear   int
	CalcuttaID       string
	TeamID           string
	SchoolName       string
	TotalBid         float64
	CalcuttaTotalBid float64
	NormalizedBid    float64
}

type BestInvestmentData struct {
	TournamentName   string
	TournamentYear   int
	CalcuttaID       string
	TeamID           string
	SchoolName       string
	Seed             int
	Region           string
	TeamPoints       float64
	TotalBid         float64
	CalcuttaTotalBid float64
	CalcuttaTotalPts float64
	InvestmentShare  float64
	PointsShare      float64
	RawROI           float64
	NormalizedROI    float64
}

type InvestmentLeaderboardData struct {
	TournamentName      string
	TournamentYear      int
	CalcuttaID          string
	EntryID             string
	EntryName           string
	TeamID              string
	SchoolName          string
	Seed                int
	Investment          float64
	OwnershipPercentage float64
	RawReturns          float64
	NormalizedReturns   float64
}

type EntryLeaderboardData struct {
	TournamentName    string
	TournamentYear    int
	CalcuttaID        string
	EntryID           string
	EntryName         string
	TotalReturns      float64
	TotalParticipants int
	AverageReturns    float64
	NormalizedReturns float64
}

type CareerLeaderboardData struct {
	EntryName           string
	Years               int
	BestFinish          int
	Wins                int
	Podiums             int
	InTheMoneys         int
	Top10s              int
	CareerEarningsCents int
}

func (r *AnalyticsRepository) GetSeedAnalytics(ctx context.Context) ([]SeedAnalyticsData, float64, float64, error) {
	query := `
		SELECT 
			tt.seed,
			COALESCE(SUM(cpt.actual_points), 0) as total_points,
			COALESCE(SUM(cet.bid), 0) as total_investment,
			COUNT(DISTINCT tt.id) as team_count
		FROM tournament_teams tt
		LEFT JOIN calcutta_entry_teams cet ON cet.team_id = tt.id AND cet.deleted_at IS NULL
		LEFT JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
		LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
		LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id AND cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
		WHERE tt.deleted_at IS NULL
		GROUP BY tt.seed
		ORDER BY tt.seed
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error executing seed analytics query: %v", err)
		return nil, 0, 0, err
	}
	defer rows.Close()

	var results []SeedAnalyticsData
	var totalPoints, totalInvestment float64

	for rows.Next() {
		var data SeedAnalyticsData
		if err := rows.Scan(&data.Seed, &data.TotalPoints, &data.TotalInvestment, &data.TeamCount); err != nil {
			log.Printf("Error scanning seed analytics row: %v", err)
			return nil, 0, 0, err
		}
		results = append(results, data)
		totalPoints += data.TotalPoints
		totalInvestment += data.TotalInvestment
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating seed analytics rows: %v", err)
		return nil, 0, 0, err
	}

	return results, totalPoints, totalInvestment, nil
}

func (r *AnalyticsRepository) GetRegionAnalytics(ctx context.Context) ([]RegionAnalyticsData, float64, float64, error) {
	query := `
		SELECT 
			tt.region,
			COALESCE(SUM(cpt.actual_points), 0) as total_points,
			COALESCE(SUM(cet.bid), 0) as total_investment,
			COUNT(DISTINCT tt.id) as team_count
		FROM tournament_teams tt
		LEFT JOIN calcutta_entry_teams cet ON cet.team_id = tt.id AND cet.deleted_at IS NULL
		LEFT JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
		LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
		LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id AND cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
		WHERE tt.deleted_at IS NULL
		GROUP BY tt.region
		ORDER BY tt.region
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error executing region analytics query: %v", err)
		return nil, 0, 0, err
	}
	defer rows.Close()

	var results []RegionAnalyticsData
	var totalPoints, totalInvestment float64

	for rows.Next() {
		var data RegionAnalyticsData
		if err := rows.Scan(&data.Region, &data.TotalPoints, &data.TotalInvestment, &data.TeamCount); err != nil {
			log.Printf("Error scanning region analytics row: %v", err)
			return nil, 0, 0, err
		}
		results = append(results, data)
		totalPoints += data.TotalPoints
		totalInvestment += data.TotalInvestment
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating region analytics rows: %v", err)
		return nil, 0, 0, err
	}

	return results, totalPoints, totalInvestment, nil
}

func (r *AnalyticsRepository) GetTeamAnalytics(ctx context.Context) ([]TeamAnalyticsData, error) {
	query := `
		SELECT 
			s.id as school_id,
			s.name as school_name,
			COALESCE(SUM(cpt.actual_points), 0) as total_points,
			COALESCE(SUM(cet.bid), 0) as total_investment,
			COUNT(DISTINCT tt.id) as appearances,
			COALESCE(SUM(tt.seed), 0) as total_seed
		FROM schools s
		LEFT JOIN tournament_teams tt ON tt.school_id = s.id AND tt.deleted_at IS NULL
		LEFT JOIN calcutta_entry_teams cet ON cet.team_id = tt.id AND cet.deleted_at IS NULL
		LEFT JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
		LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
		LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id AND cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
		WHERE s.deleted_at IS NULL
		GROUP BY s.id, s.name
		HAVING COUNT(DISTINCT tt.id) > 0
		ORDER BY total_points DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error executing team analytics query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []TeamAnalyticsData

	for rows.Next() {
		var data TeamAnalyticsData
		if err := rows.Scan(&data.SchoolID, &data.SchoolName, &data.TotalPoints, &data.TotalInvestment, &data.Appearances, &data.TotalSeed); err != nil {
			log.Printf("Error scanning team analytics row: %v", err)
			return nil, err
		}
		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating team analytics rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (r *AnalyticsRepository) GetBestCareers(ctx context.Context, limit int) ([]CareerLeaderboardData, error) {
	query := `
		WITH entry_points AS (
			SELECT
				c.id as calcutta_id,
				ce.id as entry_id,
				ce.created_at as entry_created_at,
				TRIM(ce.name) as entry_name,
				COALESCE(SUM(cpt.actual_points), 0)::float as total_points
			FROM calcutta_entries ce
			JOIN calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
			LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
			LEFT JOIN calcutta_portfolio_teams cpt ON cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
			WHERE ce.deleted_at IS NULL
			GROUP BY c.id, ce.id, ce.created_at, ce.name
		),
		group_stats AS (
			SELECT
				calcutta_id,
				total_points,
				COUNT(*)::int as tie_size,
				DENSE_RANK() OVER (PARTITION BY calcutta_id ORDER BY total_points DESC) as group_rank
			FROM entry_points
			GROUP BY calcutta_id, total_points
		),
		position_groups AS (
			SELECT
				gs.*,
				1 + COALESCE(
					SUM(tie_size) OVER (PARTITION BY calcutta_id ORDER BY group_rank ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING),
					0
				) as start_position
			FROM group_stats gs
		),
		enriched AS (
			SELECT
				ep.calcutta_id,
				ep.entry_id,
				ep.entry_name,
				ep.total_points,
				pg.tie_size,
				pg.start_position as finish_position,
				ROW_NUMBER() OVER (
					PARTITION BY ep.calcutta_id, ep.total_points
					ORDER BY ep.entry_created_at DESC, ep.entry_id ASC
				)::int as tie_index
			FROM entry_points ep
			JOIN position_groups pg
				ON pg.calcutta_id = ep.calcutta_id AND pg.total_points = ep.total_points
			WHERE ep.entry_name <> ''
		),
		payout_calc AS (
			SELECT
				e.*,
				COALESCE(gp.group_payout_cents, 0)::int as group_payout_cents,
				(COALESCE(gp.group_payout_cents, 0)::int / NULLIF(e.tie_size, 0))::int as base_payout_cents,
				(COALESCE(gp.group_payout_cents, 0)::int % NULLIF(e.tie_size, 0))::int as remainder_cents
			FROM enriched e
			LEFT JOIN LATERAL (
				SELECT COALESCE(SUM(cp.amount_cents), 0)::int as group_payout_cents
				FROM calcutta_payouts cp
				WHERE cp.calcutta_id = e.calcutta_id
					AND cp.deleted_at IS NULL
					AND cp.position >= e.finish_position
					AND cp.position < (e.finish_position + e.tie_size)
			) gp ON true
		),
		entry_results AS (
			SELECT
				pc.entry_name,
				pc.calcutta_id,
				pc.finish_position,
				(
					pc.base_payout_cents + CASE WHEN pc.tie_index <= pc.remainder_cents THEN 1 ELSE 0 END
				)::int as payout_cents
			FROM payout_calc pc
		),
		career_agg AS (
			SELECT
				entry_name,
				COUNT(DISTINCT calcutta_id)::int as years,
				MIN(finish_position)::int as best_finish,
				SUM(CASE WHEN finish_position = 1 THEN 1 ELSE 0 END)::int as wins,
				SUM(CASE WHEN finish_position <= 3 THEN 1 ELSE 0 END)::int as podiums,
				SUM(CASE WHEN payout_cents > 0 THEN 1 ELSE 0 END)::int as in_the_moneys,
				SUM(CASE WHEN finish_position <= 10 THEN 1 ELSE 0 END)::int as top_10s,
				SUM(payout_cents)::int as career_earnings_cents
			FROM entry_results
			GROUP BY entry_name
		)
		SELECT
			entry_name,
			years,
			best_finish,
			wins,
			podiums,
			in_the_moneys,
			top_10s,
			career_earnings_cents
		FROM career_agg
		ORDER BY
			(career_earnings_cents::float / NULLIF(years, 0)) DESC,
			wins DESC,
			podiums DESC,
			in_the_moneys DESC,
			top_10s DESC,
			entry_name ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		log.Printf("Error executing best careers query: %v", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]CareerLeaderboardData, 0)
	for rows.Next() {
		var d CareerLeaderboardData
		if err := rows.Scan(
			&d.EntryName,
			&d.Years,
			&d.BestFinish,
			&d.Wins,
			&d.Podiums,
			&d.InTheMoneys,
			&d.Top10s,
			&d.CareerEarningsCents,
		); err != nil {
			log.Printf("Error scanning best careers row: %v", err)
			return nil, err
		}
		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating best careers rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (r *AnalyticsRepository) GetBestInvestmentBids(ctx context.Context, limit int) ([]InvestmentLeaderboardData, error) {
	query := `
		WITH bid_points AS (
			SELECT
				t.name as tournament_name,
				COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year,
				c.id as calcutta_id,
				ce.id as entry_id,
				ce.name as entry_name,
				tt.id as team_id,
				s.name as school_name,
				tt.seed,
				CASE (tt.wins + tt.byes)
					WHEN 0 THEN 0
					WHEN 1 THEN 0
					WHEN 2 THEN 50
					WHEN 3 THEN 150
					WHEN 4 THEN 300
					WHEN 5 THEN 500
					WHEN 6 THEN 750
					WHEN 7 THEN 1050
					ELSE 0
				END::float as team_points,
				cet.bid::float as bid,
				SUM(cet.bid::float) OVER (PARTITION BY c.id, tt.id) as team_total_bid,
				SUM(cet.bid::float) OVER (PARTITION BY c.id) as calcutta_total_bid
			FROM calcutta_entry_teams cet
			JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
			JOIN calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			JOIN tournament_teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
			JOIN schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
			WHERE cet.deleted_at IS NULL
		),
		calcutta_points AS (
			SELECT
				distinct_teams.calcutta_id,
				SUM(distinct_teams.team_points) as calcutta_total_points
			FROM (
				SELECT DISTINCT calcutta_id, team_id, team_points
				FROM bid_points
			) distinct_teams
			GROUP BY distinct_teams.calcutta_id
		),
		calcutta_participants AS (
			SELECT
				bp.calcutta_id,
				COUNT(DISTINCT bp.entry_id)::float as total_participants
			FROM bid_points bp
			GROUP BY bp.calcutta_id
		)
		SELECT
			tournament_name,
			tournament_year,
			bp.calcutta_id,
			entry_id,
			entry_name,
			team_id,
			school_name,
			seed,
			bid as investment,
			CASE WHEN team_total_bid > 0 THEN (bid / team_total_bid) ELSE 0 END as ownership_percentage,
			CASE WHEN team_total_bid > 0 THEN (team_points * (bid / team_total_bid)) ELSE 0 END as raw_returns,
			CASE
				WHEN team_total_bid > 0 AND cp.calcutta_total_points > 0 AND cpar.total_participants > 0 THEN (team_points * (bid / team_total_bid)) * (100.0 * cpar.total_participants) / cp.calcutta_total_points
				ELSE 0
			END as normalized_returns
		FROM bid_points bp
		JOIN calcutta_points cp ON cp.calcutta_id = bp.calcutta_id
		JOIN calcutta_participants cpar ON cpar.calcutta_id = bp.calcutta_id
		WHERE bid > 0
			AND team_points > 0
		ORDER BY normalized_returns DESC, raw_returns DESC, investment ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		log.Printf("Error executing best investment bids query: %v", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]InvestmentLeaderboardData, 0)
	for rows.Next() {
		var d InvestmentLeaderboardData
		if err := rows.Scan(
			&d.TournamentName,
			&d.TournamentYear,
			&d.CalcuttaID,
			&d.EntryID,
			&d.EntryName,
			&d.TeamID,
			&d.SchoolName,
			&d.Seed,
			&d.Investment,
			&d.OwnershipPercentage,
			&d.RawReturns,
			&d.NormalizedReturns,
		); err != nil {
			log.Printf("Error scanning best investment bid row: %v", err)
			return nil, err
		}
		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating best investment bids rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (r *AnalyticsRepository) GetBestEntries(ctx context.Context, limit int) ([]EntryLeaderboardData, error) {
	query := `
		WITH entry_points AS (
			SELECT
				t.name as tournament_name,
				COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year,
				c.id as calcutta_id,
				ce.id as entry_id,
				ce.name as entry_name,
				COALESCE(SUM(cpt.actual_points), 0)::float as total_points
			FROM calcutta_entries ce
			JOIN calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
			LEFT JOIN calcutta_portfolio_teams cpt ON cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
			WHERE ce.deleted_at IS NULL
			GROUP BY t.name, tournament_year, c.id, ce.id, ce.name
		),
		enriched AS (
			SELECT
				entry_points.*,
				COUNT(*) OVER (PARTITION BY calcutta_id) as total_participants,
				AVG(total_points) OVER (PARTITION BY calcutta_id) as average_returns,
				SUM(total_points) OVER (PARTITION BY calcutta_id) as calcutta_total_returns
			FROM entry_points
		)
		SELECT
			tournament_name,
			tournament_year,
			calcutta_id,
			entry_id,
			entry_name,
			total_points as total_returns,
			total_participants,
			average_returns,
			CASE
				WHEN calcutta_total_returns > 0 AND total_participants > 0 THEN total_points * (100.0 * total_participants::float) / calcutta_total_returns
				ELSE 0
			END as normalized_returns
		FROM enriched
		WHERE total_points > 0
		ORDER BY normalized_returns DESC, total_points DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		log.Printf("Error executing best entries query: %v", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]EntryLeaderboardData, 0)
	for rows.Next() {
		var d EntryLeaderboardData
		if err := rows.Scan(
			&d.TournamentName,
			&d.TournamentYear,
			&d.CalcuttaID,
			&d.EntryID,
			&d.EntryName,
			&d.TotalReturns,
			&d.TotalParticipants,
			&d.AverageReturns,
			&d.NormalizedReturns,
		); err != nil {
			log.Printf("Error scanning best entries row: %v", err)
			return nil, err
		}
		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating best entries rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (r *AnalyticsRepository) GetBestInvestments(ctx context.Context, limit int) ([]BestInvestmentData, error) {
	query := `
		WITH team_bids AS (
			SELECT
				t.name as tournament_name,
				COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year,
				c.id as calcutta_id,
				tt.id as team_id,
				s.name as school_name,
				tt.seed,
				tt.region,
				CASE (tt.wins + tt.byes)
					WHEN 0 THEN 0
					WHEN 1 THEN 0
					WHEN 2 THEN 50
					WHEN 3 THEN 150
					WHEN 4 THEN 300
					WHEN 5 THEN 500
					WHEN 6 THEN 750
					WHEN 7 THEN 1050
					ELSE 0
				END::float as team_points,
				SUM(cet.bid)::float as total_bid
			FROM calcutta_entry_teams cet
			JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
			JOIN calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			JOIN tournament_teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
			JOIN schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
			WHERE cet.deleted_at IS NULL
			GROUP BY t.name, tournament_year, c.id, tt.id, s.name, tt.seed, tt.region, team_points
		),
		calcutta_enriched AS (
			SELECT
				team_bids.*,
				SUM(total_bid) OVER (PARTITION BY calcutta_id) as calcutta_total_bid,
				SUM(team_points) OVER (PARTITION BY calcutta_id) as calcutta_total_points
			FROM team_bids
		)
		SELECT
			tournament_name,
			tournament_year,
			calcutta_id,
			team_id,
			school_name,
			seed,
			region,
			team_points,
			total_bid,
			calcutta_total_bid,
			calcutta_total_points,
			CASE WHEN calcutta_total_bid > 0 THEN (total_bid / calcutta_total_bid) ELSE 0 END as investment_share,
			CASE WHEN calcutta_total_points > 0 THEN (team_points / calcutta_total_points) ELSE 0 END as points_share,
			CASE WHEN total_bid > 0 THEN (team_points / total_bid) ELSE 0 END as raw_roi,
			CASE
				WHEN total_bid > 0 AND calcutta_total_bid > 0 AND calcutta_total_points > 0 THEN (team_points / total_bid) / (calcutta_total_points / calcutta_total_bid)
				ELSE 0
			END as normalized_roi
		FROM calcutta_enriched
		WHERE total_bid > 0
			AND team_points > 0
		ORDER BY normalized_roi DESC, raw_roi DESC, total_bid ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		log.Printf("Error executing best investments query: %v", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]BestInvestmentData, 0)
	for rows.Next() {
		var d BestInvestmentData
		if err := rows.Scan(
			&d.TournamentName,
			&d.TournamentYear,
			&d.CalcuttaID,
			&d.TeamID,
			&d.SchoolName,
			&d.Seed,
			&d.Region,
			&d.TeamPoints,
			&d.TotalBid,
			&d.CalcuttaTotalBid,
			&d.CalcuttaTotalPts,
			&d.InvestmentShare,
			&d.PointsShare,
			&d.RawROI,
			&d.NormalizedROI,
		); err != nil {
			log.Printf("Error scanning best investments row: %v", err)
			return nil, err
		}
		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating best investments rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (r *AnalyticsRepository) GetSeedInvestmentPoints(ctx context.Context) ([]SeedInvestmentPointData, error) {
	query := `
		WITH team_bids AS (
			SELECT
				tt.seed,
				t.name as tournament_name,
				COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year,
				c.id as calcutta_id,
				tt.id as team_id,
				s.name as school_name,
				SUM(cet.bid)::float as total_bid
			FROM calcutta_entry_teams cet
			JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
			JOIN calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			JOIN tournament_teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
			JOIN schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
			WHERE cet.deleted_at IS NULL
				AND (tt.byes > 0 OR tt.wins > 0)
			GROUP BY tt.seed, t.name, tournament_year, c.id, tt.id, s.name
		)
		SELECT
			seed,
			tournament_name,
			tournament_year,
			calcutta_id,
			team_id,
			school_name,
			total_bid,
			SUM(total_bid) OVER (PARTITION BY calcutta_id) as calcutta_total_bid,
			CASE
				WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
				ELSE 0
			END as normalized_bid
		FROM team_bids
		ORDER BY seed, tournament_name, calcutta_id, total_bid DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error executing seed investment points query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []SeedInvestmentPointData
	for rows.Next() {
		var d SeedInvestmentPointData
		if err := rows.Scan(&d.Seed, &d.TournamentName, &d.TournamentYear, &d.CalcuttaID, &d.TeamID, &d.SchoolName, &d.TotalBid, &d.CalcuttaTotalBid, &d.NormalizedBid); err != nil {
			log.Printf("Error scanning seed investment points row: %v", err)
			return nil, err
		}
		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating seed investment points rows: %v", err)
		return nil, err
	}

	return results, nil
}

func (r *AnalyticsRepository) GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceData, error) {
	query := `
		WITH team_bids AS (
			SELECT
				tt.seed,
				c.id as calcutta_id,
				tt.id as team_id,
				SUM(cet.bid)::float as total_bid,
				MAX(cpt.actual_points) as actual_points
			FROM calcutta_entry_teams cet
			JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
			JOIN calcuttas c ON c.id = ce.calcutta_id AND c.deleted_at IS NULL
			JOIN tournament_teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
			LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
			LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id AND cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
			WHERE cet.deleted_at IS NULL
				AND (tt.byes > 0 OR tt.wins > 0)
			GROUP BY tt.seed, c.id, tt.id
		),
		normalized AS (
			SELECT
				seed,
				CASE
					WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
					ELSE 0
				END as normalized_bid,
				actual_points
			FROM team_bids
		)
		SELECT
			seed,
			COALESCE(STDDEV(normalized_bid), 0) as investment_stddev,
			COALESCE(STDDEV(actual_points), 0) as points_stddev,
			COALESCE(AVG(normalized_bid), 0) as investment_mean,
			COALESCE(AVG(actual_points), 0) as points_mean,
			COUNT(*) as team_count
		FROM normalized
		GROUP BY seed
		HAVING COUNT(*) > 1
		ORDER BY seed
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error executing seed variance analytics query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []SeedVarianceData

	for rows.Next() {
		var data SeedVarianceData
		if err := rows.Scan(&data.Seed, &data.InvestmentStdDev, &data.PointsStdDev, &data.InvestmentMean, &data.PointsMean, &data.TeamCount); err != nil {
			log.Printf("Error scanning seed variance analytics row: %v", err)
			return nil, err
		}
		results = append(results, data)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating seed variance analytics rows: %v", err)
		return nil, err
	}

	return results, nil
}
