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
	"math"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrNoCalcuttaForYear = errors.New("no calcuttas found for year")

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
	TotalCommunityBidExcl  float64
	CalcuttaTotalExcl      float64
	NormalizedBidExcl      float64
}

func dpAllocateBids(predPoints []float64, baseBids []float64, budget int, minTeams int, maxTeams int, minBid int, maxBid int) ([]int, int, error) {
	n := len(predPoints)
	if n != len(baseBids) {
		return nil, 0, fmt.Errorf("predPoints and baseBids length mismatch")
	}
	negInf := math.Inf(-1)

	// dp[k][b] = best value with k teams selected and total bid b.
	dp := make([][]float64, maxTeams+1)
	for k := 0; k <= maxTeams; k++ {
		dp[k] = make([]float64, budget+1)
		for b := 0; b <= budget; b++ {
			dp[k][b] = negInf
		}
	}
	dp[0][0] = 0

	type parent struct {
		prevK  int
		prevB  int
		choice int
	}
	parents := make([][][]parent, n+1)
	for i := 0; i <= n; i++ {
		parents[i] = make([][]parent, maxTeams+1)
		for k := 0; k <= maxTeams; k++ {
			parents[i][k] = make([]parent, budget+1)
		}
	}

	for i := 0; i < n; i++ {
		newDP := make([][]float64, maxTeams+1)
		for k := 0; k <= maxTeams; k++ {
			newDP[k] = make([]float64, budget+1)
			for b := 0; b <= budget; b++ {
				newDP[k][b] = negInf
			}
		}

		for k := 0; k <= maxTeams; k++ {
			for b := 0; b <= budget; b++ {
				cur := dp[k][b]
				if cur == negInf {
					continue
				}

				// Choice x = 0 (skip team i)
				if cur > newDP[k][b] {
					newDP[k][b] = cur
					parents[i+1][k][b] = parent{prevK: k, prevB: b, choice: 0}
				}

				// Choices x in [minBid..maxBid]
				if k == maxTeams {
					continue
				}
				for x := minBid; x <= maxBid; x++ {
					if b+x > budget {
						break
					}
					den := baseBids[i] + float64(x)
					val := 0.0
					if den > 0 {
						val = predPoints[i] * float64(x) / den
					}
					next := cur + val
					if next > newDP[k+1][b+x] {
						newDP[k+1][b+x] = next
						parents[i+1][k+1][b+x] = parent{prevK: k, prevB: b, choice: x}
					}
				}
			}
		}

		dp = newDP
	}

	bestK := -1
	bestVal := negInf
	for k := minTeams; k <= maxTeams; k++ {
		v := dp[k][budget]
		if v > bestVal {
			bestVal = v
			bestK = k
		}
	}
	if bestK == -1 || bestVal == negInf {
		return nil, 0, fmt.Errorf("no feasible allocation found")
	}

	bids := make([]int, n)
	k := bestK
	b := budget
	for i := n; i >= 1; i-- {
		p := parents[i][k][b]
		bids[i-1] = p.choice
		k = p.prevK
		b = p.prevB
	}

	selected := 0
	for _, x := range bids {
		if x > 0 {
			selected++
		}
	}

	return bids, selected, nil
}

type SimulateRow struct {
	TournamentName      string
	TournamentYear      int
	CalcuttaID          string
	TeamID              string
	SchoolName          string
	Seed                int
	Region              string
	BaseMarketBid       float64
	BaseMarketBidShare  float64
	PredPoints          float64
	PredPointsShare     float64
	RecommendedBid      int
	OwnershipAfter      float64
	ExpectedEntryPoints float64
}

func main() {
	var (
		mode             = flag.String("mode", "export", "Mode to run: export|baseline|simulate|backtest")
		year             = flag.Int("year", 0, "Tournament year to export (matches 4-digit year parsed from tournament name).")
		calcuttaID       = flag.String("calcutta-id", "", "Calcutta ID to export.")
		outPath          = flag.String("out", "", "Output path for CSV (defaults to stdout).")
		trainYears       = flag.Int("train-years", 0, "Baseline training window size in prior years (e.g. 1,2,3). 0 means use all available history excluding the target calcutta.")
		excludeEntryName = flag.String("exclude-entry-name", "", "Exclude bids from entries with this name (e.g. to reduce strategy leakage / measure cannibalization).")
		budget           = flag.Int("budget", 100, "Total budget to allocate for simulate mode.")
		minTeams         = flag.Int("min-teams", 3, "Minimum number of teams to bid on in simulate mode.")
		maxTeams         = flag.Int("max-teams", 10, "Maximum number of teams to bid on in simulate mode.")
		minBid           = flag.Int("min-bid", 1, "Minimum bid per team in simulate mode.")
		maxBid           = flag.Int("max-bid", 50, "Maximum bid per team in simulate mode.")
		startYear        = flag.Int("start-year", 0, "Start year for backtest mode.")
		endYear          = flag.Int("end-year", 0, "End year for backtest mode.")
	)
	flag.Parse()

	if *mode != "backtest" {
		if *year == 0 && *calcuttaID == "" {
			log.Fatal("Must provide either -year or -calcutta-id")
		}
		if *year != 0 && *calcuttaID != "" {
			log.Fatal("Provide only one of -year or -calcutta-id")
		}
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
		rows, err := queryTeamDataset(ctx, db, *year, *calcuttaID, *excludeEntryName)
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
		rows, summary, err := runSeedBaseline(ctx, db, *calcuttaID, *trainYears, *excludeEntryName)
		if err != nil {
			log.Fatalf("Failed to run baseline: %v", err)
		}
		log.Printf("Baseline summary: train_years=%d exclude_entry_name=%q points_mae=%.4f bid_share_mae=%.6f", *trainYears, *excludeEntryName, summary.PointsMAE, summary.BidShareMAE)
		if err := writeBaselineCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "simulate":
		if *calcuttaID == "" {
			log.Fatal("simulate mode requires -calcutta-id (or -year that resolves to a single calcutta)")
		}
		rows, summary, err := runSimulateEntry(ctx, db, *calcuttaID, *trainYears, *excludeEntryName, *budget, *minTeams, *maxTeams, *minBid, *maxBid)
		if err != nil {
			log.Fatalf("Failed to simulate entry: %v", err)
		}
		log.Printf("Simulate summary: train_years=%d exclude_entry_name=%q teams=%d budget=%d expected_points_share=%.6f expected_bid_share=%.6f expected_normalized_roi=%.4f", *trainYears, *excludeEntryName, summary.NumTeams, summary.Budget, summary.ExpectedPointsShare, summary.ExpectedBidShare, summary.ExpectedNormalizedROI)
		if err := writeSimulateCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "backtest":
		if *startYear == 0 || *endYear == 0 {
			log.Fatal("backtest mode requires -start-year and -end-year")
		}
		if *endYear < *startYear {
			log.Fatal("backtest mode requires -end-year >= -start-year")
		}
		rows, err := runBacktest(ctx, db, *startYear, *endYear, *trainYears, *excludeEntryName, *budget, *minTeams, *maxTeams, *minBid, *maxBid)
		if err != nil {
			log.Fatalf("Failed to run backtest: %v", err)
		}
		if err := writeBacktestCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	default:
		log.Fatalf("Unknown -mode %q (expected export|baseline|simulate|backtest)", *mode)
	}
}

type BaselineSummary struct {
	PointsMAE   float64
	BidShareMAE float64
}

type SimulateSummary struct {
	Budget                int
	NumTeams              int
	ExpectedPointsShare   float64
	ExpectedBidShare      float64
	ExpectedNormalizedROI float64
}

type BacktestRow struct {
	TournamentYear        int
	CalcuttaID            string
	TrainYears            int
	ExcludeEntryName      string
	NumTeams              int
	Budget                int
	ExpectedPointsShare   float64
	ExpectedBidShare      float64
	ExpectedNormalizedROI float64
	RealizedPointsShare   float64
	RealizedBidShare      float64
	RealizedNormalizedROI float64
}

type BaselineRow struct {
	TournamentName     string
	TournamentYear     int
	CalcuttaID         string
	TeamID             string
	SchoolName         string
	Seed               int
	Region             string
	ActualPoints       float64
	PredPoints         float64
	ActualBidShare     float64
	ActualBidShareExcl float64
	PredBidShare       float64
	ActualROI          float64
	PredROI            float64
}

func runSeedBaseline(ctx context.Context, db *sql.DB, targetCalcuttaID string, trainYears int, excludeEntryName string) ([]BaselineRow, *BaselineSummary, error) {
	targetRows, err := queryTeamDataset(ctx, db, 0, targetCalcuttaID, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}

	targetYear, err := calcuttaYear(ctx, db, targetCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	maxYear := targetYear - 1
	minYear := 0
	if trainYears > 0 {
		minYear = targetYear - trainYears
	}
	if trainYears > 0 && maxYear < minYear {
		return nil, nil, fmt.Errorf("invalid training window: target_year=%d train_years=%d", targetYear, trainYears)
	}

	seedPointsMean, seedBidShareMean, err := computeSeedMeans(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}
	if len(seedPointsMean) == 0 {
		return nil, nil, fmt.Errorf("no training data found for baseline: target_year=%d train_years=%d", targetYear, trainYears)
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

		actualBidShareUsed := r.NormalizedBid
		if excludeEntryName != "" {
			actualBidShareUsed = r.NormalizedBidExcl
		}

		var actualPointsShare float64
		if totalActualPoints > 0 {
			actualPointsShare = r.TeamPoints / totalActualPoints
		}
		var predPointsShare float64
		if totalPredPoints > 0 {
			predPointsShare = predPoints / totalPredPoints
		}

		var actualROI float64
		if actualBidShareUsed > 0 {
			actualROI = actualPointsShare / actualBidShareUsed
		}
		var predROI float64
		if predBidShare > 0 {
			predROI = predPointsShare / predBidShare
		}

		absPointsErrSum += absFloat64(predPoints - r.TeamPoints)
		absBidShareErrSum += absFloat64(predBidShare - actualBidShareUsed)

		baselineRows = append(baselineRows, BaselineRow{
			TournamentName:     r.TournamentName,
			TournamentYear:     r.TournamentYear,
			CalcuttaID:         r.CalcuttaID,
			TeamID:             r.TeamID,
			SchoolName:         r.SchoolName,
			Seed:               r.Seed,
			Region:             r.Region,
			ActualPoints:       r.TeamPoints,
			PredPoints:         predPoints,
			ActualBidShare:     r.NormalizedBid,
			ActualBidShareExcl: r.NormalizedBidExcl,
			PredBidShare:       predBidShare,
			ActualROI:          actualROI,
			PredROI:            predROI,
		})
	}

	summary := &BaselineSummary{}
	if len(targetRows) > 0 {
		summary.PointsMAE = absPointsErrSum / float64(len(targetRows))
		summary.BidShareMAE = absBidShareErrSum / float64(len(targetRows))
	}

	return baselineRows, summary, nil
}

func runSimulateEntry(ctx context.Context, db *sql.DB, targetCalcuttaID string, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int) ([]SimulateRow, *SimulateSummary, error) {
	if budget <= 0 {
		return nil, nil, fmt.Errorf("budget must be > 0")
	}
	if minTeams <= 0 || maxTeams <= 0 || minTeams > maxTeams {
		return nil, nil, fmt.Errorf("invalid team constraints: min_teams=%d max_teams=%d", minTeams, maxTeams)
	}
	if minBid <= 0 || maxBid <= 0 || minBid > maxBid {
		return nil, nil, fmt.Errorf("invalid bid constraints: min_bid=%d max_bid=%d", minBid, maxBid)
	}
	if budget < minTeams*minBid {
		return nil, nil, fmt.Errorf("budget too small to satisfy minimums: budget=%d min_teams=%d min_bid=%d", budget, minTeams, minBid)
	}
	if budget > maxTeams*maxBid {
		return nil, nil, fmt.Errorf("budget too large to satisfy maximums: budget=%d max_teams=%d max_bid=%d", budget, maxTeams, maxBid)
	}

	targetRows, err := queryTeamDataset(ctx, db, 0, targetCalcuttaID, excludeEntryName)
	if err != nil {
		return nil, nil, err
	}

	targetYear, err := calcuttaYear(ctx, db, targetCalcuttaID)
	if err != nil {
		return nil, nil, err
	}

	maxYear := targetYear - 1
	minYear := 0
	if trainYears > 0 {
		minYear = targetYear - trainYears
	}
	if trainYears > 0 && maxYear < minYear {
		return nil, nil, fmt.Errorf("invalid training window: target_year=%d train_years=%d", targetYear, trainYears)
	}

	seedPointsMean, _, err := computeSeedMeans(ctx, db, targetCalcuttaID, trainYears, minYear, maxYear, "")
	if err != nil {
		return nil, nil, err
	}
	if len(seedPointsMean) == 0 {
		return nil, nil, fmt.Errorf("no training data found for simulate: target_year=%d train_years=%d", targetYear, trainYears)
	}

	totalPredPoints := 0.0
	for _, r := range targetRows {
		totalPredPoints += seedPointsMean[r.Seed]
	}

	totalMarketBid := 0.0
	if len(targetRows) > 0 {
		if excludeEntryName != "" {
			totalMarketBid = targetRows[0].CalcuttaTotalExcl
		} else {
			totalMarketBid = targetRows[0].CalcuttaTotalCommunity
		}
	}

	type candidate struct {
		idx        int
		predPoints float64
		baseBid    float64
		bid        int
		seed       int
	}

	candidates := make([]candidate, 0, len(targetRows))
	for i, r := range targetRows {
		baseBid := r.TotalCommunityBid
		if excludeEntryName != "" {
			baseBid = r.TotalCommunityBidExcl
		}
		candidates = append(candidates, candidate{
			idx:        i,
			predPoints: seedPointsMean[r.Seed],
			baseBid:    baseBid,
			bid:        0,
			seed:       r.Seed,
		})
	}

	predPointsVec := make([]float64, len(candidates))
	baseBidsVec := make([]float64, len(candidates))
	for i := range candidates {
		predPointsVec[i] = candidates[i].predPoints
		baseBidsVec[i] = candidates[i].baseBid
	}

	bids, selected, err := dpAllocateBids(predPointsVec, baseBidsVec, budget, minTeams, maxTeams, minBid, maxBid)
	if err != nil {
		return nil, nil, err
	}
	for i := range candidates {
		candidates[i].bid = bids[i]
	}

	byIdx := make([]candidate, len(candidates))
	for _, c := range candidates {
		byIdx[c.idx] = c
	}

	simRows := make([]SimulateRow, 0, selected)
	expectedEntryPointsTotal := 0.0
	for i, r := range targetRows {
		c := byIdx[i]
		if c.bid <= 0 {
			continue
		}
		baseBid := c.baseBid
		predPoints := c.predPoints
		predPointsShare := 0.0
		if totalPredPoints > 0 {
			predPointsShare = predPoints / totalPredPoints
		}
		baseShare := 0.0
		if excludeEntryName != "" {
			baseShare = r.NormalizedBidExcl
		} else {
			baseShare = r.NormalizedBid
		}
		ownership := 0.0
		if c.bid > 0 {
			ownership = float64(c.bid) / (baseBid + float64(c.bid))
		}
		expectedEntryPoints := predPoints * ownership
		expectedEntryPointsTotal += expectedEntryPoints
		simRows = append(simRows, SimulateRow{
			TournamentName:      r.TournamentName,
			TournamentYear:      r.TournamentYear,
			CalcuttaID:          r.CalcuttaID,
			TeamID:              r.TeamID,
			SchoolName:          r.SchoolName,
			Seed:                r.Seed,
			Region:              r.Region,
			BaseMarketBid:       baseBid,
			BaseMarketBidShare:  baseShare,
			PredPoints:          predPoints,
			PredPointsShare:     predPointsShare,
			RecommendedBid:      c.bid,
			OwnershipAfter:      ownership,
			ExpectedEntryPoints: expectedEntryPoints,
		})
	}

	pointsShare := 0.0
	if totalPredPoints > 0 {
		pointsShare = expectedEntryPointsTotal / totalPredPoints
	}
	bidShare := 0.0
	if totalMarketBid+float64(budget) > 0 {
		bidShare = float64(budget) / (totalMarketBid + float64(budget))
	}
	normROI := 0.0
	if bidShare > 0 {
		normROI = pointsShare / bidShare
	}

	summary := &SimulateSummary{
		Budget:                budget,
		NumTeams:              selected,
		ExpectedPointsShare:   pointsShare,
		ExpectedBidShare:      bidShare,
		ExpectedNormalizedROI: normROI,
	}

	return simRows, summary, nil
}

func runBacktest(ctx context.Context, db *sql.DB, startYear int, endYear int, trainYears int, excludeEntryName string, budget int, minTeams int, maxTeams int, minBid int, maxBid int) ([]BacktestRow, error) {
	rows := make([]BacktestRow, 0)
	for y := startYear; y <= endYear; y++ {
		calcuttaID, err := resolveSingleCalcuttaIDForYear(ctx, db, y)
		if err != nil {
			if errors.Is(err, ErrNoCalcuttaForYear) {
				continue
			}
			return nil, err
		}

		simRows, simSummary, err := runSimulateEntry(ctx, db, calcuttaID, trainYears, excludeEntryName, budget, minTeams, maxTeams, minBid, maxBid)
		if err != nil {
			return nil, err
		}

		datasetRows, err := queryTeamDataset(ctx, db, 0, calcuttaID, excludeEntryName)
		if err != nil {
			return nil, err
		}

		teamPointsByID := map[string]float64{}
		totalActualPoints := 0.0
		for _, r := range datasetRows {
			teamPointsByID[r.TeamID] = r.TeamPoints
			totalActualPoints += r.TeamPoints
		}

		realizedEntryPointsTotal := 0.0
		for _, r := range simRows {
			teamPoints := teamPointsByID[r.TeamID]
			own := 0.0
			if r.RecommendedBid > 0 {
				own = float64(r.RecommendedBid) / (r.BaseMarketBid + float64(r.RecommendedBid))
			}
			realizedEntryPointsTotal += teamPoints * own
		}

		realizedPointsShare := 0.0
		if totalActualPoints > 0 {
			realizedPointsShare = realizedEntryPointsTotal / totalActualPoints
		}
		realizedBidShare := simSummary.ExpectedBidShare
		realizedNormROI := 0.0
		if realizedBidShare > 0 {
			realizedNormROI = realizedPointsShare / realizedBidShare
		}

		rows = append(rows, BacktestRow{
			TournamentYear:        y,
			CalcuttaID:            calcuttaID,
			TrainYears:            trainYears,
			ExcludeEntryName:      excludeEntryName,
			NumTeams:              simSummary.NumTeams,
			Budget:                simSummary.Budget,
			ExpectedPointsShare:   simSummary.ExpectedPointsShare,
			ExpectedBidShare:      simSummary.ExpectedBidShare,
			ExpectedNormalizedROI: simSummary.ExpectedNormalizedROI,
			RealizedPointsShare:   realizedPointsShare,
			RealizedBidShare:      realizedBidShare,
			RealizedNormalizedROI: realizedNormROI,
		})
	}

	return rows, nil
}

func calcuttaYear(ctx context.Context, db *sql.DB, calcuttaID string) (int, error) {
	query := `
		SELECT COALESCE(substring(t.name from '([0-9]{4})')::int, 0) as tournament_year
		FROM calcuttas c
		JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
			AND c.id = $1::uuid
	`

	var year int
	if err := db.QueryRowContext(ctx, query, calcuttaID).Scan(&year); err != nil {
		return 0, err
	}
	if year == 0 {
		return 0, fmt.Errorf("failed to parse tournament year for calcutta-id %s", calcuttaID)
	}
	return year, nil
}

func computeSeedMeans(ctx context.Context, db *sql.DB, excludeCalcuttaID string, trainYears int, minYear int, maxYear int, excludeEntryName string) (map[int]float64, map[int]float64, error) {
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
				COALESCE(SUM(
					CASE
						WHEN $5 <> '' AND ce.name = $5 THEN 0
						ELSE COALESCE(cet.bid, 0)
					END
				), 0)::float as total_bid
			FROM tournament_teams tt
			JOIN calcuttas c ON c.tournament_id = tt.tournament_id AND c.deleted_at IS NULL
			JOIN tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
			LEFT JOIN calcutta_entries ce ON ce.calcutta_id = c.id AND ce.deleted_at IS NULL
			LEFT JOIN calcutta_entry_teams cet ON cet.entry_id = ce.id AND cet.team_id = tt.id AND cet.deleted_at IS NULL
			WHERE tt.deleted_at IS NULL
				AND c.id <> $1::uuid
				AND (
					$2 = 0
					OR
					(
						COALESCE(substring(t.name from '([0-9]{4})')::int, 0) <> 0
						AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) >= $3
						AND COALESCE(substring(t.name from '([0-9]{4})')::int, 0) <= $4
					)
				)
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

	rows, err := db.QueryContext(ctx, query, excludeCalcuttaID, trainYears, minYear, maxYear, excludeEntryName)
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
		return "", fmt.Errorf("%w: %d", ErrNoCalcuttaForYear, year)
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

func queryTeamDataset(ctx context.Context, db *sql.DB, year int, calcuttaID string, excludeEntryName string) ([]TeamDatasetRow, error) {
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
				COALESCE(SUM(COALESCE(cet.bid, 0)), 0)::float as total_bid,
				COALESCE(SUM(
					CASE
						WHEN $3 <> '' AND ce.name = $3 THEN 0
						ELSE COALESCE(cet.bid, 0)
					END
				), 0)::float as total_bid_excl
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
			END as normalized_bid,
			total_bid_excl,
			SUM(total_bid_excl) OVER (PARTITION BY calcutta_id) as calcutta_total_bid_excl,
			CASE
				WHEN SUM(total_bid_excl) OVER (PARTITION BY calcutta_id) > 0 THEN (total_bid_excl / SUM(total_bid_excl) OVER (PARTITION BY calcutta_id))
				ELSE 0
			END as normalized_bid_excl
		FROM team_bids
		ORDER BY seed ASC, total_bid DESC, school_name ASC
	`

	dbRows, err := db.QueryContext(ctx, query, calcuttaParam, year, excludeEntryName)
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
			&r.TotalCommunityBidExcl,
			&r.CalcuttaTotalExcl,
			&r.NormalizedBidExcl,
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
		"actual_bid_share_excl",
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
			fmt.Sprintf("%g", r.ActualBidShareExcl),
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

func writeSimulateCSV(w io.Writer, rows []SimulateRow) error {
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
		"base_market_bid",
		"base_market_bid_share",
		"pred_points",
		"pred_points_share",
		"recommended_bid",
		"ownership_after",
		"expected_entry_points",
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
			fmt.Sprintf("%g", r.BaseMarketBid),
			fmt.Sprintf("%g", r.BaseMarketBidShare),
			fmt.Sprintf("%g", r.PredPoints),
			fmt.Sprintf("%g", r.PredPointsShare),
			fmt.Sprintf("%d", r.RecommendedBid),
			fmt.Sprintf("%g", r.OwnershipAfter),
			fmt.Sprintf("%g", r.ExpectedEntryPoints),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

func writeBacktestCSV(w io.Writer, rows []BacktestRow) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"tournament_year",
		"calcutta_id",
		"train_years",
		"exclude_entry_name",
		"num_teams",
		"budget",
		"expected_points_share",
		"expected_bid_share",
		"expected_normalized_roi",
		"realized_points_share",
		"realized_bid_share",
		"realized_normalized_roi",
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		record := []string{
			fmt.Sprintf("%d", r.TournamentYear),
			r.CalcuttaID,
			fmt.Sprintf("%d", r.TrainYears),
			r.ExcludeEntryName,
			fmt.Sprintf("%d", r.NumTeams),
			fmt.Sprintf("%d", r.Budget),
			fmt.Sprintf("%g", r.ExpectedPointsShare),
			fmt.Sprintf("%g", r.ExpectedBidShare),
			fmt.Sprintf("%g", r.ExpectedNormalizedROI),
			fmt.Sprintf("%g", r.RealizedPointsShare),
			fmt.Sprintf("%g", r.RealizedBidShare),
			fmt.Sprintf("%g", r.RealizedNormalizedROI),
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
		"total_community_investment_excl",
		"calcutta_total_investment_excl",
		"normalized_bid_excl",
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
			fmt.Sprintf("%g", r.TotalCommunityBidExcl),
			fmt.Sprintf("%g", r.CalcuttaTotalExcl),
			fmt.Sprintf("%g", r.NormalizedBidExcl),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
