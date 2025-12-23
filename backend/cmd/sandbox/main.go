package main

import (
	"context"
	"database/sql"
	"flag"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	var (
		mode             = flag.String("mode", "export", "Mode to run: export|baseline|simulate|backtest|report|kenpom-returns|invest-eval")
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
		predModel        = flag.String("pred-model", "seed", "Predicted returns model to use for simulate/report/backtest: seed|kenpom")
		investModel      = flag.String("invest-model", "seed", "Predicted investment model to use for baseline evaluation: seed|seed-pod|seed-kenpom-delta|seed-kenpom-rank|kenpom-rank|kenpom-score")
		sigma            = flag.Float64("sigma", 11.0, "Sigma (std dev) for KenPom margin->win probability conversion.")
	)
	flag.Parse()

	if *mode != "backtest" && *mode != "report" && *mode != "invest-eval" {
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
	case "kenpom-returns":
		if *calcuttaID == "" {
			log.Fatal("kenpom-returns mode requires -calcutta-id (or -year that resolves to a single calcutta)")
		}
		rows, err := runKenPomReturns(ctx, db, *calcuttaID, *sigma)
		if err != nil {
			log.Fatalf("Failed to compute kenpom returns: %v", err)
		}
		if err := writeKenPomReturnsCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "baseline":
		if *calcuttaID == "" {
			log.Fatal("baseline mode requires -calcutta-id (or -year that resolves to a single calcutta)")
		}
		rows, summary, err := runSeedBaseline(ctx, db, *calcuttaID, *trainYears, *excludeEntryName, *investModel)
		if err != nil {
			log.Fatalf("Failed to run baseline: %v", err)
		}
		log.Printf("Baseline summary: invest_model=%s train_years=%d exclude_entry_name=%q points_mae=%.4f bid_share_mae=%.6f", *investModel, *trainYears, *excludeEntryName, summary.PointsMAE, summary.BidShareMAE)
		if err := writeBaselineCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "simulate":
		if *calcuttaID == "" {
			log.Fatal("simulate mode requires -calcutta-id (or -year that resolves to a single calcutta)")
		}
		rows, summary, err := runSimulateEntry(ctx, db, *calcuttaID, *trainYears, *excludeEntryName, *budget, *minTeams, *maxTeams, *minBid, *maxBid, *predModel, *investModel, *sigma)
		if err != nil {
			log.Fatalf("Failed to simulate entry: %v", err)
		}
		log.Printf("Simulate summary: pred_model=%s sigma=%g train_years=%d exclude_entry_name=%q teams=%d budget=%d expected_points_share=%.6f expected_bid_share=%.6f expected_normalized_roi=%.4f", *predModel, *sigma, *trainYears, *excludeEntryName, summary.NumTeams, summary.Budget, summary.ExpectedPointsShare, summary.ExpectedBidShare, summary.ExpectedNormalizedROI)
		if err := writeSimulateCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "backtest":
		start := *startYear
		end := *endYear
		if start == 0 || end == 0 {
			minY, maxY, err := availableTournamentYearRange(ctx, db)
			if err != nil {
				log.Fatalf("Failed to determine available year range: %v", err)
			}
			if start == 0 {
				start = minY
			}
			if end == 0 {
				end = maxY
			}
		}
		if end < start {
			log.Fatal("backtest mode requires -end-year >= -start-year")
		}
		rows, err := runBacktest(ctx, db, start, end, *trainYears, *excludeEntryName, *budget, *minTeams, *maxTeams, *minBid, *maxBid, *predModel, *investModel, *sigma)
		if err != nil {
			log.Fatalf("Failed to run backtest: %v", err)
		}
		if err := writeBacktestCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	case "report":
		start := *startYear
		end := *endYear
		if start == 0 || end == 0 {
			minY, maxY, err := availableTournamentYearRange(ctx, db)
			if err != nil {
				log.Fatalf("Failed to determine available year range: %v", err)
			}
			if start == 0 {
				start = minY
			}
			if end == 0 {
				end = maxY
			}
		}
		if end < start {
			log.Fatal("report mode requires -end-year >= -start-year")
		}
		if err := runReport(ctx, db, out, start, end, *trainYears, *excludeEntryName, *budget, *minTeams, *maxTeams, *minBid, *maxBid, *predModel, *investModel, *sigma); err != nil {
			log.Fatalf("Failed to run report: %v", err)
		}
	case "invest-eval":
		start := *startYear
		end := *endYear
		if start == 0 || end == 0 {
			minY, maxY, err := availableTournamentYearRange(ctx, db)
			if err != nil {
				log.Fatalf("Failed to determine available year range: %v", err)
			}
			if start == 0 {
				start = minY
			}
			if end == 0 {
				end = maxY
			}
		}
		if end < start {
			log.Fatal("invest-eval mode requires -end-year >= -start-year")
		}
		rows, err := runInvestEval(ctx, db, start, end, *trainYears, *excludeEntryName)
		if err != nil {
			log.Fatalf("Failed to run invest-eval: %v", err)
		}
		if err := writeInvestEvalCSV(out, rows); err != nil {
			log.Fatalf("Failed to write CSV: %v", err)
		}
	default:
		log.Fatalf("Unknown -mode %q (expected export|baseline|simulate|backtest|report|kenpom-returns|invest-eval)", *mode)
	}
}
