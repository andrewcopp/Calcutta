package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
)

type InvestEvalRow struct {
	TournamentYear               int
	CalcuttaID                   string
	TrainYears                   int
	ExcludeEntry                 string
	InvestModel                  string
	Teams                        int
	BidShareMAE                  float64
	BidShareMAEEqualSeed         float64
	BidShareMAESeedNorm          float64
	BidShareResidMAE             float64
	BidShareResidMAEStd          float64
	UnderbidTopKHitRate          float64
	OverbidTopKHitRate           float64
	UnderbidTopKHitRateTop16     float64
	OverbidTopKHitRateTop16      float64
	UnderbidTopKHitRateSeeds1To8 float64
	OverbidTopKHitRateSeeds1To8  float64
	BidShareMSE                  float64
	BidShareRMSE                 float64
}

func runInvestEval(ctx context.Context, db *sql.DB, startYear int, endYear int, trainYears int, excludeEntryName string, models []string) ([]InvestEvalRow, error) {
	if len(models) == 0 {
		models = []string{"seed", "seed-pod", "seed-kenpom-delta", "seed-kenpom-rank", "kenpom-rank", "kenpom-score", "ridge", "ridge-returns"}
	}
	type agg struct {
		sse                float64
		sae                float64
		seedNormAbsSum     float64
		residAbsSum        float64
		residStdAbsSum     float64
		underHits          int
		overHits           int
		underN             int
		overN              int
		underHitsTop16     int
		overHitsTop16      int
		underNTop16        int
		overNTop16         int
		underHitsSeeds1To8 int
		overHitsSeeds1To8  int
		underNSeeds1To8    int
		overNSeeds1To8     int
		n                  int
	}
	overall := make(map[string]agg, len(models))
	overallBySeed := make(map[string]map[int]floatAgg, len(models))
	overallResidBySeed := make(map[string]map[int]floatAgg, len(models))

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

			// Compute training-window seed mean/std for bid share to stabilize normalized metrics.
			maxYear := y - 1
			minYear := 0
			if trainYears > 0 {
				minYear = y - trainYears
			}
			trainRows, err := queryTrainingBidShares(ctx, db, calcuttaID, trainYears, minYear, maxYear, excludeEntryName)
			if err != nil {
				return nil, err
			}
			trainMeanBySeed := map[int]float64{}
			trainStdBySeed := map[int]float64{}
			if len(trainRows) > 0 {
				seedAgg := map[int]floatAgg{}
				seedSqAgg := map[int]floatAgg{}
				for _, tr := range trainRows {
					a := seedAgg[tr.Seed]
					a.Sum += tr.BidShare
					a.Count++
					seedAgg[tr.Seed] = a
					b := seedSqAgg[tr.Seed]
					b.Sum += tr.BidShare * tr.BidShare
					b.Count++
					seedSqAgg[tr.Seed] = b
				}
				for seed, a := range seedAgg {
					if a.Count == 0 {
						continue
					}
					mu := a.Sum / float64(a.Count)
					trainMeanBySeed[seed] = mu
					sq := seedSqAgg[seed]
					if sq.Count > 0 {
						v := sq.Sum/float64(sq.Count) - mu*mu
						if v < 0 {
							v = 0
						}
						trainStdBySeed[seed] = math.Sqrt(v)
					}
				}
			}

			actualByTeam := make(map[string]float64, len(datasetRows))
			seedByTeam := make(map[string]int, len(datasetRows))
			for _, r := range datasetRows {
				actual := r.NormalizedBid
				if excludeEntryName != "" {
					actual = r.NormalizedBidExcl
				}
				actualByTeam[r.TeamID] = actual
				seedByTeam[r.TeamID] = r.Seed
			}
			meanActualBySeed := meanBySeed(actualByTeam, seedByTeam)
			meanPredBySeed := meanBySeed(predShareByTeam, seedByTeam)
			stdActualBySeed := stddevBySeed(actualByTeam, seedByTeam, meanActualBySeed)

			sse := 0.0
			sae := 0.0
			seedNormAbsSum := 0.0
			residAbsSum := 0.0
			residStdAbsSum := 0.0
			n := 0

			bySeed := overallBySeed[m]
			if bySeed == nil {
				bySeed = map[int]floatAgg{}
				overallBySeed[m] = bySeed
			}
			residBySeed := overallResidBySeed[m]
			if residBySeed == nil {
				residBySeed = map[int]floatAgg{}
				overallResidBySeed[m] = residBySeed
			}

			for _, r := range datasetRows {
				actual := actualByTeam[r.TeamID]
				pred := predShareByTeam[r.TeamID]
				d := pred - actual
				sae += absFloat64(d)
				sse += d * d

				denom := trainMeanBySeed[r.Seed]
				if denom == 0 {
					denom = meanActualBySeed[r.Seed]
				}
				if denom < 1e-9 {
					denom = 1e-9
				}
				seedNormAbsSum += absFloat64(d) / denom

				residActual := actual - meanActualBySeed[r.Seed]
				residPred := pred - meanPredBySeed[r.Seed]
				dr := residPred - residActual
				residAbsSum += absFloat64(dr)
				std := trainStdBySeed[r.Seed]
				if std == 0 {
					std = stdActualBySeed[r.Seed]
				}
				if std < 1e-9 {
					std = 1e-9
				}
				residStdAbsSum += absFloat64(dr) / std
				n++

				aSeed := bySeed[r.Seed]
				aSeed.Sum += absFloat64(d)
				aSeed.Count++
				bySeed[r.Seed] = aSeed

				aResid := residBySeed[r.Seed]
				aResid.Sum += absFloat64(dr)
				aResid.Count++
				residBySeed[r.Seed] = aResid
			}
			if n == 0 {
				continue
			}
			mae := sae / float64(n)
			maeSeedNorm := seedNormAbsSum / float64(n)
			residMAE := residAbsSum / float64(n)
			residMAEStd := residStdAbsSum / float64(n)
			mse := sse / float64(n)
			rmse := math.Sqrt(mse)
			maeEqualSeed := equalWeightBySeedMAE(actualByTeam, predShareByTeam, seedByTeam)
			k := 3
			underHits, overHits, usedK := topKResidualHitCounts(actualByTeam, predShareByTeam, seedByTeam, meanActualBySeed, meanPredBySeed, k)
			underHitRate := 0.0
			overHitRate := 0.0
			if usedK > 0 {
				underHitRate = float64(underHits) / float64(usedK)
				overHitRate = float64(overHits) / float64(usedK)
			}

			allowedSeeds1To8 := map[string]struct{}{}
			for _, r := range datasetRows {
				if r.Seed >= 1 && r.Seed <= 8 {
					allowedSeeds1To8[r.TeamID] = struct{}{}
				}
			}

			type kpItem struct {
				TeamID string
				KP     float64
			}
			kpItems := make([]kpItem, 0, len(datasetRows))
			for _, r := range datasetRows {
				v := math.Inf(-1)
				if r.KenPomNetRtg != nil {
					v = *r.KenPomNetRtg
				}
				kpItems = append(kpItems, kpItem{TeamID: r.TeamID, KP: v})
			}
			sort.Slice(kpItems, func(i, j int) bool {
				if kpItems[i].KP == kpItems[j].KP {
					return kpItems[i].TeamID < kpItems[j].TeamID
				}
				return kpItems[i].KP > kpItems[j].KP
			})
			allowedTop16 := map[string]struct{}{}
			for i := 0; i < len(kpItems) && i < 16; i++ {
				allowedTop16[kpItems[i].TeamID] = struct{}{}
			}

			underHitsTop16, overHitsTop16, usedKTop16 := topKResidualHitCountsAllowed(actualByTeam, predShareByTeam, seedByTeam, meanActualBySeed, meanPredBySeed, k, allowedTop16)
			underHitRateTop16 := 0.0
			overHitRateTop16 := 0.0
			if usedKTop16 > 0 {
				underHitRateTop16 = float64(underHitsTop16) / float64(usedKTop16)
				overHitRateTop16 = float64(overHitsTop16) / float64(usedKTop16)
			}

			underHitsSeeds1To8, overHitsSeeds1To8, usedKSeeds1To8 := topKResidualHitCountsAllowed(actualByTeam, predShareByTeam, seedByTeam, meanActualBySeed, meanPredBySeed, k, allowedSeeds1To8)
			underHitRateSeeds1To8 := 0.0
			overHitRateSeeds1To8 := 0.0
			if usedKSeeds1To8 > 0 {
				underHitRateSeeds1To8 = float64(underHitsSeeds1To8) / float64(usedKSeeds1To8)
				overHitRateSeeds1To8 = float64(overHitsSeeds1To8) / float64(usedKSeeds1To8)
			}

			a := overall[m]
			a.sse += sse
			a.sae += sae
			a.seedNormAbsSum += seedNormAbsSum
			a.residAbsSum += residAbsSum
			a.residStdAbsSum += residStdAbsSum
			a.underHits += underHits
			a.overHits += overHits
			a.underN += usedK
			a.overN += usedK
			a.underHitsTop16 += underHitsTop16
			a.overHitsTop16 += overHitsTop16
			a.underNTop16 += usedKTop16
			a.overNTop16 += usedKTop16
			a.underHitsSeeds1To8 += underHitsSeeds1To8
			a.overHitsSeeds1To8 += overHitsSeeds1To8
			a.underNSeeds1To8 += usedKSeeds1To8
			a.overNSeeds1To8 += usedKSeeds1To8
			a.n += n
			overall[m] = a

			rows = append(rows, InvestEvalRow{
				TournamentYear:               y,
				CalcuttaID:                   calcuttaID,
				TrainYears:                   trainYears,
				ExcludeEntry:                 excludeEntryName,
				InvestModel:                  m,
				Teams:                        n,
				BidShareMAE:                  mae,
				BidShareMAEEqualSeed:         maeEqualSeed,
				BidShareMAESeedNorm:          maeSeedNorm,
				BidShareResidMAE:             residMAE,
				BidShareResidMAEStd:          residMAEStd,
				UnderbidTopKHitRate:          underHitRate,
				OverbidTopKHitRate:           overHitRate,
				UnderbidTopKHitRateTop16:     underHitRateTop16,
				OverbidTopKHitRateTop16:      overHitRateTop16,
				UnderbidTopKHitRateSeeds1To8: underHitRateSeeds1To8,
				OverbidTopKHitRateSeeds1To8:  overHitRateSeeds1To8,
				BidShareMSE:                  mse,
				BidShareRMSE:                 rmse,
			})
		}
	}

	type overallRow struct {
		model string
		row   InvestEvalRow
	}
	overallRows := make([]overallRow, 0, len(models))
	for _, m := range models {
		a := overall[m]
		if a.n == 0 {
			continue
		}
		mae := a.sae / float64(a.n)
		maeSeedNorm := a.seedNormAbsSum / float64(a.n)
		residMAE := a.residAbsSum / float64(a.n)
		residMAEStd := a.residStdAbsSum / float64(a.n)
		mse := a.sse / float64(a.n)
		rmse := math.Sqrt(mse)
		maeEqualSeed := overallEqualWeightBySeedMAE(overallBySeed[m])
		underHit := 0.0
		overHit := 0.0
		if a.underN > 0 {
			underHit = float64(a.underHits) / float64(a.underN)
		}
		if a.overN > 0 {
			overHit = float64(a.overHits) / float64(a.overN)
		}
		underHitTop16 := 0.0
		overHitTop16 := 0.0
		if a.underNTop16 > 0 {
			underHitTop16 = float64(a.underHitsTop16) / float64(a.underNTop16)
		}
		if a.overNTop16 > 0 {
			overHitTop16 = float64(a.overHitsTop16) / float64(a.overNTop16)
		}
		underHitSeeds1To8 := 0.0
		overHitSeeds1To8 := 0.0
		if a.underNSeeds1To8 > 0 {
			underHitSeeds1To8 = float64(a.underHitsSeeds1To8) / float64(a.underNSeeds1To8)
		}
		if a.overNSeeds1To8 > 0 {
			overHitSeeds1To8 = float64(a.overHitsSeeds1To8) / float64(a.overNSeeds1To8)
		}
		overallRows = append(overallRows, overallRow{
			model: m,
			row: InvestEvalRow{
				TournamentYear:               0,
				CalcuttaID:                   "",
				TrainYears:                   trainYears,
				ExcludeEntry:                 excludeEntryName,
				InvestModel:                  m,
				Teams:                        a.n,
				BidShareMAE:                  mae,
				BidShareMAEEqualSeed:         maeEqualSeed,
				BidShareMAESeedNorm:          maeSeedNorm,
				BidShareResidMAE:             residMAE,
				BidShareResidMAEStd:          residMAEStd,
				UnderbidTopKHitRate:          underHit,
				OverbidTopKHitRate:           overHit,
				UnderbidTopKHitRateTop16:     underHitTop16,
				OverbidTopKHitRateTop16:      overHitTop16,
				UnderbidTopKHitRateSeeds1To8: underHitSeeds1To8,
				OverbidTopKHitRateSeeds1To8:  overHitSeeds1To8,
				BidShareMSE:                  mse,
				BidShareRMSE:                 rmse,
			},
		})
	}

	sort.Slice(overallRows, func(i, j int) bool {
		return overallRows[i].row.BidShareResidMAEStd < overallRows[j].row.BidShareResidMAEStd
	})
	for _, r := range overallRows {
		rows = append(rows, r.row)
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
		"bid_share_mae",
		"bid_share_mae_equal_seed",
		"bid_share_mae_seed_norm",
		"bid_share_resid_mae",
		"bid_share_resid_mae_std",
		"underbid_top3_hit_rate",
		"overbid_top3_hit_rate",
		"underbid_top3_hit_rate_top16",
		"overbid_top3_hit_rate_top16",
		"underbid_top3_hit_rate_seeds_1_to_8",
		"overbid_top3_hit_rate_seeds_1_to_8",
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
			fmt.Sprintf("%g", r.BidShareMAE),
			fmt.Sprintf("%g", r.BidShareMAEEqualSeed),
			fmt.Sprintf("%g", r.BidShareMAESeedNorm),
			fmt.Sprintf("%g", r.BidShareResidMAE),
			fmt.Sprintf("%g", r.BidShareResidMAEStd),
			fmt.Sprintf("%g", r.UnderbidTopKHitRate),
			fmt.Sprintf("%g", r.OverbidTopKHitRate),
			fmt.Sprintf("%g", r.UnderbidTopKHitRateTop16),
			fmt.Sprintf("%g", r.OverbidTopKHitRateTop16),
			fmt.Sprintf("%g", r.UnderbidTopKHitRateSeeds1To8),
			fmt.Sprintf("%g", r.OverbidTopKHitRateSeeds1To8),
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
