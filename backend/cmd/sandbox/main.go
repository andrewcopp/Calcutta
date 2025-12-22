package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type TeamDatasetRow struct {
	TournamentName         string
	TournamentYear         int
	CalcuttaID             string
	TeamID                 string
	SchoolName             string
	Seed                   int
	Region                 string
	Wins                   int
	Byes                   int
	TeamPoints             float64
	TotalCommunityBid      float64
	CalcuttaTotalCommunity float64
	NormalizedBid          float64
}

func main() {
	var (
		mode       = flag.String("mode", "export", "Mode to run: export|baseline")
		year       = flag.Int("year", 0, "Tournament year to export (matches 4-digit year parsed from tournament name).")
		calcuttaID = flag.String("calcutta-id", "", "Calcutta ID to export.")
		outPath    = flag.String("out", "", "Output path for CSV (defaults to stdout).")
	)
	flag.Parse()

	if *year == 0 && *calcuttaID == "" {
		log.Fatal("Must provide either -year or -calcutta-id")
	}
	if *year != 0 && *calcuttaID != "" {
		log.Fatal("Provide only one of -year or -calcutta-id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if *year != 0 {
		resolvedCalcuttaID, err := resolveSingleCalcuttaIDForYear(ctx, db, *year)
		if err != nil {
			log.Fatalf("Failed to resolve calcutta for year %d: %v", *year, err)
		}
		*calcuttaID = resolvedCalcuttaID
		*year = 0
	}

	var out io.Writer = os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer func() {
			_ = f.Close()
		}()
		out = f
	}

	switch *mode {
	case "export":
		rows, err := queryTeamDataset(ctx, db, *year, *calcuttaID)
		if err != nil {
			log.Fatalf("Failed to query dataset: %v", err)
		}
		if err := writeCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "baseline":
		if *calcuttaID == "" {
			log.Fatal("baseline mode requires -calcutta-id (or -year that resolves to a single calcutta)")
		}
		rows, summary, err := runSeedBaseline(ctx, db, *calcuttaID)
		if err != nil {
			log.Fatalf("Failed to run baseline: %v", err)
		}
		log.Printf("Baseline summary: points_mae=%.4f bid_share_mae=%.6f", summary.PointsMAE, summary.BidShareMAE)
		if err := writeBaselineCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	default:
		log.Fatalf("Unknown -mode %q (expected export|baseline)", *mode)
	}
}

type BaselineSummary struct {
	PointsMAE   float64
	BidShareMAE float64
}

type BaselineRow struct {
	TournamentName string
	TournamentYear int
	CalcuttaID     string
	TeamID         string
	SchoolName     string
	Seed           int
	Region         string
	ActualPoints   float64
	PredPoints     float64
	ActualBidShare float64
	PredBidShare   float64
	ActualROI      float64
	PredROI        float64
}

func runSeedBaseline(ctx context.Context, db *sql.DB, targetCalcuttaID string) ([]BaselineRow, *BaselineSummary, error) {
	targetRows, err := queryTeamDataset(ctx, db, 0, targetCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	seedPointsMean, seedBidShareMean, err := computeSeedMeans(ctx, db, targetCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	var totalActualPoints float64
	var totalPredPoints float64
	for _, r := range targetRows {
		totalActualPoints += r.TeamPoints
		totalPredPoints += seedPointsMean[r.Seed]
	}

	baselineRows := make([]BaselineRow, 0, len(targetRows))
	var absPointsErrSum float64
	var absBidShareErrSum float64

	for _, r := range targetRows {
		predPoints := seedPointsMean[r.Seed]
		predBidShare := seedBidShareMean[r.Seed]

		var actualPointsShare float64
		if totalActualPoints > 0 {
			actualPointsShare = r.TeamPoints / totalActualPoints
		}
		var predPointsShare float64
		if totalPredPoints > 0 {
			predPointsShare = predPoints / totalPredPoints
		}

		var actualROI float64
		if r.NormalizedBid > 0 {
			actualROI = actualPointsShare / r.NormalizedBid
		}
		var predROI float64
		if predBidShare > 0 {
			predROI = predPointsShare / predBidShare
		}

		absPointsErrSum += absFloat64(predPoints - r.TeamPoints)
		absBidShareErrSum += absFloat64(predBidShare - r.NormalizedBid)

		baselineRows = append(baselineRows, BaselineRow{
			TournamentName: r.TournamentName,
			TournamentYear: r.TournamentYear,
			CalcuttaID:     r.CalcuttaID,
			TeamID:         r.TeamID,
			SchoolName:     r.SchoolName,
			Seed:           r.Seed,
			Region:         r.Region,
			ActualPoints:   r.TeamPoints,
			PredPoints:     predPoints,
			ActualBidShare: r.NormalizedBid,
			PredBidShare:   predBidShare,
			ActualROI:      actualROI,
			PredROI:        predROI,
		})
	}

	summary := &BaselineSummary{}
	if len(targetRows) > 0 {
		summary.PointsMAE = absPointsErrSum / float64(len(targetRows))
		summary.BidShareMAE = absBidShareErrSum / float64(len(targetRows))
	}

	return baselineRows, summary, nil
}

func computeSeedMeans(ctx context.Context, db *sql.DB, excludeCalcuttaID string) (map[int]float64, map[int]float64, error) {
	query := `
		WITH team_bids AS (
			SELECT
				c.id as calcutta_id,
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
				COALESCE(SUM(cet.bid), 0)::float as total_bid
			FROM tournament_teams tt
			JOIN calcuttas c ON c.tournament_id = tt.tournament_id AND c.deleted_at IS NULL
			LEFT JOIN calcutta_entries ce ON ce.calcutta_id = c.id AND ce.deleted_at IS NULL
			LEFT JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.team_id = tt.id AND cet.deleted_at IS NULL
			WHERE tt.deleted_at IS NULL
				AND c.id <> $1::uuid
			GROUP BY c.id, tt.id, tt.seed, team_points
		),
		enriched AS (
			SELECT
				calcutta_id,
				seed,
				team_points,
				CASE
					WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
					ELSE 0
				END as bid_share
			FROM team_bids
		)
		SELECT
			seed,
			AVG(team_points) as mean_points,
			AVG(bid_share) as mean_bid_share
		FROM enriched
		GROUP BY seed
		ORDER BY seed
	`

	rows, err := db.QueryContext(ctx, query, excludeCalcuttaID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	seedPoints := map[int]float64{}
	seedBidShare := map[int]float64{}
	for rows.Next() {
		var seed int
		var meanPoints, meanBidShare float64
		if err := rows.Scan(&seed, &meanPoints, &meanBidShare); err != nil {
			return nil, nil, err
		}
		seedPoints[seed] = meanPoints
		seedBidShare[seed] = meanBidShare
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return seedPoints, seedBidShare, nil
}

func absFloat64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func resolveSingleCalcuttaIDForYear(ctx context.Context, db *sql.DB, year int) (string, error) {
	query := `
		SELECT c.id, c.name
		FROM calcuttas c
		JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) = $1
		ORDER BY c.created_at ASC, c.id ASC
	`

	rows, err := db.QueryContext(ctx, query, year)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type calcuttaRow struct {
		ID   string
		Name string
	}

	calcuttas := make([]calcuttaRow, 0)
	for rows.Next() {
		var r calcuttaRow
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return "", err
		}
		calcuttas = append(calcuttas, r)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	if len(calcuttas) == 0 {
		return "", fmt.Errorf("no calcuttas found for year %d", year)
	}
	if len(calcuttas) > 1 {
		msg := fmt.Sprintf("found %d calcuttas for year %d; please re-run with -calcutta-id. Candidates:\n", len(calcuttas), year)
		for _, c := range calcuttas {
			msg += fmt.Sprintf("- %s (%s)\n", c.ID, c.Name)
		}
		return "", errors.New(msg)
	}

	return calcuttas[0].ID, nil
}

func queryTeamDataset(ctx context.Context, db *sql.DB, year int, calcuttaID string) ([]TeamDatasetRow, error) {
	if year != 0 && calcuttaID != "" {
		return nil, errors.New("cannot provide both year and calcuttaID")
	}
	if year == 0 && calcuttaID == "" {
		return nil, errors.New("must provide either year or calcuttaID")
	}

	var calcuttaParam any
	if calcuttaID != "" {
		calcuttaParam = calcuttaID
	} else {
		calcuttaParam = nil
	}
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
				tt.wins,
				tt.byes,
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
				COALESCE(SUM(cet.bid), 0)::float as total_bid
			FROM tournament_teams tt
			JOIN schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
			JOIN calcuttas c ON c.tournament_id = tt.tournament_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			LEFT JOIN calcutta_entries ce ON ce.calcutta_id = c.id AND ce.deleted_at IS NULL
			LEFT JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.team_id = tt.id AND cet.deleted_at IS NULL
			WHERE tt.deleted_at IS NULL
				AND (
					(c.id = $1::uuid)
					OR
					($2 <> 0 AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) = $2)
				)
			GROUP BY
				t.name,
				tournament_year,
				c.id,
				tt.id,
				s.name,
				tt.seed,
				tt.region,
				tt.wins,
				tt.byes,
				team_points
		)
		SELECT
			tournament_name,
			tournament_year,
			calcutta_id,
			team_id,
			school_name,
			seed,
			region,
			wins,
			byes,
			team_points,
			total_bid,
			SUM(total_bid) OVER (PARTITION BY calcutta_id) as calcutta_total_bid,
			CASE
				WHEN SUM(total_bid) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid / SUM(total_bid) OVER (PARTITION BY calcutta_id))
				ELSE 0
			END as normalized_bid
		FROM team_bids
		ORDER BY seed ASC, total_bid DESC, school_name ASC
	`

	dbRows, err := db.QueryContext(ctx, query, calcuttaParam, year)
	if err != nil {
		return nil, err
	}
	defer dbRows.Close()

	results := make([]TeamDatasetRow, 0)
	for dbRows.Next() {
		var r TeamDatasetRow
		if err := dbRows.Scan(
			&r.TournamentName,
			&r.TournamentYear,
			&r.CalcuttaID,
			&r.TeamID,
			&r.SchoolName,
			&r.Seed,
			&r.Region,
			&r.Wins,
			&r.Byes,
			&r.TeamPoints,
			&r.TotalCommunityBid,
			&r.CalcuttaTotalCommunity,
			&r.NormalizedBid,
		); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	if err := dbRows.Err(); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		if year != 0 {
			return nil, fmt.Errorf("no rows returned for year %d", year)
		}
		return nil, fmt.Errorf("no rows returned for calcutta-id %s", calcuttaID)
	}

	return results, nil
}

func writeBaselineCSV(w io.Writer, rows []BaselineRow) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"tournament_name",
		"tournament_year",
		"calcutta_id",
		"team_id",
		"school_name",
		"seed",
		"region",
		"actual_points",
		"pred_points",
		"actual_bid_share",
		"pred_bid_share",
		"actual_normalized_roi",
		"pred_normalized_roi",
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		record := []string{
			r.TournamentName,
			fmt.Sprintf("%d", r.TournamentYear),
			r.CalcuttaID,
			r.TeamID,
			r.SchoolName,
			fmt.Sprintf("%d", r.Seed),
			r.Region,
			fmt.Sprintf("%g", r.ActualPoints),
			fmt.Sprintf("%g", r.PredPoints),
			fmt.Sprintf("%g", r.ActualBidShare),
			fmt.Sprintf("%g", r.PredBidShare),
			fmt.Sprintf("%g", r.ActualROI),
			fmt.Sprintf("%g", r.PredROI),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

func writeCSV(w io.Writer, rows []TeamDatasetRow) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"tournament_name",
		"tournament_year",
		"calcutta_id",
		"team_id",
		"school_name",
		"seed",
		"region",
		"wins",
		"byes",
		"team_points",
		"total_community_investment",
		"calcutta_total_investment",
		"normalized_bid",
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		record := []string{
			r.TournamentName,
			fmt.Sprintf("%d", r.TournamentYear),
			r.CalcuttaID,
			r.TeamID,
			r.SchoolName,
			fmt.Sprintf("%d", r.Seed),
			r.Region,
			fmt.Sprintf("%d", r.Wins),
			fmt.Sprintf("%d", r.Byes),
			fmt.Sprintf("%g", r.TeamPoints),
			fmt.Sprintf("%g", r.TotalCommunityBid),
			fmt.Sprintf("%g", r.CalcuttaTotalCommunity),
			fmt.Sprintf("%g", r.NormalizedBid),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
