package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
)

type InvestEvalRow struct {
	TournamentYear int
	CalcuttaID     string
	TrainYears     int
	ExcludeEntry   string
	InvestModel    string
	Teams          int
	BidShareMSE    float64
	BidShareRMSE   float64
}

func runInvestEval(ctx context.Context, db *sql.DB, startYear int, endYear int, trainYears int, excludeEntryName string) ([]InvestEvalRow, error) {
	models := []string{"seed", "seed-pod", "seed-kenpom-delta", "seed-kenpom-rank", "kenpom-rank", "kenpom-score"}
	type agg struct {
		sse float64
		n   int
	}
	overall := make(map[string]agg, len(models))

	rows := make([]InvestEvalRow, 0)
	for y := startYear; y <= endYear; y++ {
		calcuttaID, err := resolveSingleCalcuttaIDForYear(ctx, db, y)
		if err != nil {
			if errors.Is(err, ErrNoCalcuttaForYear) {
				continue
			}
			return nil, err
		}

		datasetRows, err := queryTeamDataset(ctx, db, 0, calcuttaID, excludeEntryName)
		if err != nil {
			return nil, err
		}

		for _, m := range models {
			_, predShareByTeam, _, err := predictedMarketBidsByTeam(ctx, db, calcuttaID, datasetRows, trainYears, m, excludeEntryName)
			if err != nil {
				if errors.Is(err, ErrNoTrainingData) {
					continue
				}
				return nil, err
			}

			sse := 0.0
			n := 0
			for _, r := range datasetRows {
				actual := r.NormalizedBid
				if excludeEntryName != "" {
					actual = r.NormalizedBidExcl
				}
				pred := predShareByTeam[r.TeamID]
				d := pred - actual
				sse += d * d
				n++
			}
			if n == 0 {
				continue
			}
			mse := sse / float64(n)
			rmse := math.Sqrt(mse)

			a := overall[m]
			a.sse += sse
			a.n += n
			overall[m] = a

			rows = append(rows, InvestEvalRow{
				TournamentYear: y,
				CalcuttaID:     calcuttaID,
				TrainYears:     trainYears,
				ExcludeEntry:   excludeEntryName,
				InvestModel:    m,
				Teams:          n,
				BidShareMSE:    mse,
				BidShareRMSE:   rmse,
			})
		}
	}

	for _, m := range models {
		a := overall[m]
		if a.n == 0 {
			continue
		}
		mse := a.sse / float64(a.n)
		rmse := math.Sqrt(mse)
		rows = append(rows, InvestEvalRow{
			TournamentYear: 0,
			CalcuttaID:     "",
			TrainYears:     trainYears,
			ExcludeEntry:   excludeEntryName,
			InvestModel:    m,
			Teams:          a.n,
			BidShareMSE:    mse,
			BidShareRMSE:   rmse,
		})
	}

	return rows, nil
}

func writeInvestEvalCSV(w io.Writer, rows []InvestEvalRow) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"tournament_year",
		"calcutta_id",
		"train_years",
		"exclude_entry_name",
		"invest_model",
		"teams",
		"bid_share_mse",
		"bid_share_rmse",
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	for _, r := range rows {
		rec := []string{
			fmt.Sprintf("%d", r.TournamentYear),
			r.CalcuttaID,
			fmt.Sprintf("%d", r.TrainYears),
			r.ExcludeEntry,
			r.InvestModel,
			fmt.Sprintf("%d", r.Teams),
			fmt.Sprintf("%g", r.BidShareMSE),
			fmt.Sprintf("%g", r.BidShareRMSE),
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
