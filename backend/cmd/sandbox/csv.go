package main

import (
	"encoding/csv"
	"fmt"
	"io"
)

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
