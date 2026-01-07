package dtos

import analytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"

func ToAnalyticsResponse(result *analytics.AnalyticsResult) AnalyticsResponse {
	if result == nil {
		return AnalyticsResponse{}
	}

	out := AnalyticsResponse{
		TotalPoints:     result.TotalPoints,
		TotalInvestment: result.TotalInvestment,
		BaselineROI:     result.BaselineROI,
	}

	out.SeedAnalytics = make([]SeedAnalytics, len(result.SeedAnalytics))
	for i, sa := range result.SeedAnalytics {
		out.SeedAnalytics[i] = SeedAnalytics{
			Seed:                 sa.Seed,
			TotalPoints:          sa.TotalPoints,
			TotalInvestment:      sa.TotalInvestment,
			PointsPercentage:     sa.PointsPercentage,
			InvestmentPercentage: sa.InvestmentPercentage,
			TeamCount:            sa.TeamCount,
			AveragePoints:        sa.AveragePoints,
			AverageInvestment:    sa.AverageInvestment,
			ROI:                  sa.ROI,
		}
	}

	out.RegionAnalytics = make([]RegionAnalytics, len(result.RegionAnalytics))
	for i, ra := range result.RegionAnalytics {
		out.RegionAnalytics[i] = RegionAnalytics{
			Region:               ra.Region,
			TotalPoints:          ra.TotalPoints,
			TotalInvestment:      ra.TotalInvestment,
			PointsPercentage:     ra.PointsPercentage,
			InvestmentPercentage: ra.InvestmentPercentage,
			TeamCount:            ra.TeamCount,
			AveragePoints:        ra.AveragePoints,
			AverageInvestment:    ra.AverageInvestment,
			ROI:                  ra.ROI,
		}
	}

	out.TeamAnalytics = make([]TeamAnalytics, len(result.TeamAnalytics))
	for i, ta := range result.TeamAnalytics {
		out.TeamAnalytics[i] = TeamAnalytics{
			SchoolID:          ta.SchoolID,
			SchoolName:        ta.SchoolName,
			TotalPoints:       ta.TotalPoints,
			TotalInvestment:   ta.TotalInvestment,
			Appearances:       ta.Appearances,
			AveragePoints:     ta.AveragePoints,
			AverageInvestment: ta.AverageInvestment,
			AverageSeed:       ta.AverageSeed,
			ROI:               ta.ROI,
		}
	}

	out.SeedVarianceAnalytics = make([]SeedVarianceAnalytics, len(result.SeedVarianceAnalytics))
	for i, sv := range result.SeedVarianceAnalytics {
		out.SeedVarianceAnalytics[i] = SeedVarianceAnalytics{
			Seed:             sv.Seed,
			InvestmentStdDev: sv.InvestmentStdDev,
			PointsStdDev:     sv.PointsStdDev,
			InvestmentMean:   sv.InvestmentMean,
			PointsMean:       sv.PointsMean,
			InvestmentCV:     sv.InvestmentCV,
			PointsCV:         sv.PointsCV,
			TeamCount:        sv.TeamCount,
			VarianceRatio:    sv.VarianceRatio,
		}
	}

	return out
}

func ToSeedInvestmentDistributionResponse(result *analytics.SeedInvestmentDistributionResult) SeedInvestmentDistributionResponse {
	if result == nil {
		return SeedInvestmentDistributionResponse{}
	}

	out := SeedInvestmentDistributionResponse{}

	out.Points = make([]SeedInvestmentPoint, 0, len(result.Points))
	for _, p := range result.Points {
		out.Points = append(out.Points, SeedInvestmentPoint{
			Seed:             p.Seed,
			TournamentName:   p.TournamentName,
			TournamentYear:   p.TournamentYear,
			CalcuttaID:       p.CalcuttaID,
			TeamID:           p.TeamID,
			SchoolName:       p.SchoolName,
			TotalBid:         p.TotalBid,
			CalcuttaTotalBid: p.CalcuttaTotalBid,
			NormalizedBid:    p.NormalizedBid,
		})
	}

	out.Summaries = make([]SeedInvestmentSummary, 0, len(result.Summaries))
	for _, s := range result.Summaries {
		out.Summaries = append(out.Summaries, SeedInvestmentSummary{
			Seed:   s.Seed,
			Count:  s.Count,
			Mean:   s.Mean,
			StdDev: s.StdDev,
			Min:    s.Min,
			Q1:     s.Q1,
			Median: s.Median,
			Q3:     s.Q3,
			Max:    s.Max,
		})
	}

	return out
}

func ToBestTeamsResponse(results []analytics.BestInvestmentResult) BestTeamsResponse {
	out := BestTeamsResponse{Teams: make([]BestTeam, 0, len(results))}
	for _, bi := range results {
		out.Teams = append(out.Teams, BestTeam{
			TournamentName:   bi.TournamentName,
			TournamentYear:   bi.TournamentYear,
			CalcuttaID:       bi.CalcuttaID,
			TeamID:           bi.TeamID,
			SchoolName:       bi.SchoolName,
			Seed:             bi.Seed,
			Region:           bi.Region,
			TeamPoints:       bi.TeamPoints,
			TotalBid:         bi.TotalBid,
			CalcuttaTotalBid: bi.CalcuttaTotalBid,
			CalcuttaTotalPts: bi.CalcuttaTotalPts,
			InvestmentShare:  bi.InvestmentShare,
			PointsShare:      bi.PointsShare,
			RawROI:           bi.RawROI,
			NormalizedROI:    bi.NormalizedROI,
		})
	}
	return out
}

func ToBestInvestmentsResponse(results []analytics.BestInvestmentResult) BestInvestmentsResponse {
	out := BestInvestmentsResponse{Investments: make([]BestInvestment, 0, len(results))}
	for _, bi := range results {
		out.Investments = append(out.Investments, BestInvestment{
			TournamentName:   bi.TournamentName,
			TournamentYear:   bi.TournamentYear,
			CalcuttaID:       bi.CalcuttaID,
			TeamID:           bi.TeamID,
			SchoolName:       bi.SchoolName,
			Seed:             bi.Seed,
			Region:           bi.Region,
			TeamPoints:       bi.TeamPoints,
			TotalBid:         bi.TotalBid,
			CalcuttaTotalBid: bi.CalcuttaTotalBid,
			CalcuttaTotalPts: bi.CalcuttaTotalPts,
			InvestmentShare:  bi.InvestmentShare,
			PointsShare:      bi.PointsShare,
			RawROI:           bi.RawROI,
			NormalizedROI:    bi.NormalizedROI,
		})
	}
	return out
}

func ToInvestmentLeaderboardResponse(results []analytics.InvestmentLeaderboardResult) InvestmentLeaderboardResponse {
	out := InvestmentLeaderboardResponse{Investments: make([]InvestmentLeaderboardRow, 0, len(results))}
	for _, inv := range results {
		out.Investments = append(out.Investments, InvestmentLeaderboardRow{
			TournamentName:      inv.TournamentName,
			TournamentYear:      inv.TournamentYear,
			CalcuttaID:          inv.CalcuttaID,
			EntryID:             inv.EntryID,
			EntryName:           inv.EntryName,
			TeamID:              inv.TeamID,
			SchoolName:          inv.SchoolName,
			Seed:                inv.Seed,
			Investment:          inv.Investment,
			OwnershipPercentage: inv.OwnershipPercentage,
			RawReturns:          inv.RawReturns,
			NormalizedReturns:   inv.NormalizedReturns,
		})
	}
	return out
}

func ToEntryLeaderboardResponse(results []analytics.EntryLeaderboardResult) EntryLeaderboardResponse {
	out := EntryLeaderboardResponse{Entries: make([]EntryLeaderboardRow, 0, len(results))}
	for _, e := range results {
		out.Entries = append(out.Entries, EntryLeaderboardRow{
			TournamentName:    e.TournamentName,
			TournamentYear:    e.TournamentYear,
			CalcuttaID:        e.CalcuttaID,
			EntryID:           e.EntryID,
			EntryName:         e.EntryName,
			TotalReturns:      e.TotalReturns,
			TotalParticipants: e.TotalParticipants,
			AverageReturns:    e.AverageReturns,
			NormalizedReturns: e.NormalizedReturns,
		})
	}
	return out
}

func ToCareerLeaderboardResponse(results []analytics.CareerLeaderboardResult) CareerLeaderboardResponse {
	out := CareerLeaderboardResponse{Careers: make([]CareerLeaderboardRow, 0, len(results))}
	for _, c := range results {
		out.Careers = append(out.Careers, CareerLeaderboardRow{
			EntryName:              c.EntryName,
			Years:                  c.Years,
			BestFinish:             c.BestFinish,
			Wins:                   c.Wins,
			Podiums:                c.Podiums,
			InTheMoneys:            c.InTheMoneys,
			Top10s:                 c.Top10s,
			CareerEarningsCents:    c.CareerEarningsCents,
			ActiveInLatestCalcutta: c.ActiveInLatestCalcutta,
		})
	}
	return out
}
