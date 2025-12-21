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

func (r *AnalyticsRepository) GetSeedVarianceAnalytics(ctx context.Context) ([]SeedVarianceData, error) {
	query := `
		SELECT 
			tt.seed,
			COALESCE(STDDEV(cet.bid), 0) as investment_stddev,
			COALESCE(STDDEV(cpt.actual_points), 0) as points_stddev,
			COALESCE(AVG(cet.bid), 0) as investment_mean,
			COALESCE(AVG(cpt.actual_points), 0) as points_mean,
			COUNT(DISTINCT tt.id) as team_count
		FROM tournament_teams tt
		LEFT JOIN calcutta_entry_teams cet ON cet.team_id = tt.id AND cet.deleted_at IS NULL
		LEFT JOIN calcutta_entries ce ON ce.id = cet.entry_id AND ce.deleted_at IS NULL
		LEFT JOIN calcutta_portfolios cp ON cp.entry_id = ce.id AND cp.deleted_at IS NULL
		LEFT JOIN calcutta_portfolio_teams cpt ON cpt.team_id = tt.id AND cpt.portfolio_id = cp.id AND cpt.deleted_at IS NULL
		WHERE tt.deleted_at IS NULL
		GROUP BY tt.seed
		HAVING COUNT(DISTINCT tt.id) > 1
		ORDER BY tt.seed
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
