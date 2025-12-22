package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/andrewcopp/Calcutta/backend/cmd/server/dtos"
)

func (s *Server) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := s.analyticsService.GetAllAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting analytics: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.AnalyticsResponse{
		TotalPoints:     result.TotalPoints,
		TotalInvestment: result.TotalInvestment,
		BaselineROI:     result.BaselineROI,
	}

	response.SeedAnalytics = make([]dtos.SeedAnalytics, len(result.SeedAnalytics))
	for i, sa := range result.SeedAnalytics {
		response.SeedAnalytics[i] = dtos.SeedAnalytics{
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

	response.RegionAnalytics = make([]dtos.RegionAnalytics, len(result.RegionAnalytics))
	for i, ra := range result.RegionAnalytics {
		response.RegionAnalytics[i] = dtos.RegionAnalytics{
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

	response.TeamAnalytics = make([]dtos.TeamAnalytics, len(result.TeamAnalytics))
	for i, ta := range result.TeamAnalytics {
		response.TeamAnalytics[i] = dtos.TeamAnalytics{
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

	response.SeedVarianceAnalytics = make([]dtos.SeedVarianceAnalytics, len(result.SeedVarianceAnalytics))
	for i, sv := range result.SeedVarianceAnalytics {
		response.SeedVarianceAnalytics[i] = dtos.SeedVarianceAnalytics{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestTeamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	results, err := s.analyticsService.GetBestInvestments(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best teams: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.BestTeamsResponse{Teams: make([]dtos.BestTeam, 0, len(results))}
	for _, bi := range results {
		response.Teams = append(response.Teams, dtos.BestTeam{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	results, err := s.analyticsService.GetBestInvestmentBids(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best investments: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.InvestmentLeaderboardResponse{Investments: make([]dtos.InvestmentLeaderboardRow, 0, len(results))}
	for _, inv := range results {
		response.Investments = append(response.Investments, dtos.InvestmentLeaderboardRow{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestEntriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	results, err := s.analyticsService.GetBestEntries(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best entries: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.EntryLeaderboardResponse{Entries: make([]dtos.EntryLeaderboardRow, 0, len(results))}
	for _, e := range results {
		response.Entries = append(response.Entries, dtos.EntryLeaderboardRow{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestCareersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	results, err := s.analyticsService.GetBestCareers(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best careers: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.CareerLeaderboardResponse{Careers: make([]dtos.CareerLeaderboardRow, 0, len(results))}
	for _, c := range results {
		response.Careers = append(response.Careers, dtos.CareerLeaderboardRow{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) bestInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			limit = v
		}
	}

	results, err := s.analyticsService.GetBestInvestments(ctx, limit)
	if err != nil {
		log.Printf("Error getting best investments: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.BestInvestmentsResponse{Investments: make([]dtos.BestInvestment, 0, len(results))}
	for _, bi := range results {
		response.Investments = append(response.Investments, dtos.BestInvestment{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) seedInvestmentDistributionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	distribution, err := s.analyticsService.GetSeedInvestmentDistribution(ctx)
	if err != nil {
		log.Printf("Error getting seed investment distribution: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.SeedInvestmentDistributionResponse{}

	response.Points = make([]dtos.SeedInvestmentPoint, 0, len(distribution.Points))
	for _, p := range distribution.Points {
		response.Points = append(response.Points, dtos.SeedInvestmentPoint{
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

	response.Summaries = make([]dtos.SeedInvestmentSummary, 0, len(distribution.Summaries))
	for _, s := range distribution.Summaries {
		response.Summaries = append(response.Summaries, dtos.SeedInvestmentSummary{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) seedAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seedAnalytics, totalPoints, totalInvestment, err := s.analyticsService.GetSeedAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting seed analytics: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	response := dtos.AnalyticsResponse{
		TotalPoints:     totalPoints,
		TotalInvestment: totalInvestment,
		BaselineROI:     baselineROI,
	}

	response.SeedAnalytics = make([]dtos.SeedAnalytics, len(seedAnalytics))
	for i, sa := range seedAnalytics {
		response.SeedAnalytics[i] = dtos.SeedAnalytics{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) regionAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	regionAnalytics, totalPoints, totalInvestment, err := s.analyticsService.GetRegionAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting region analytics: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	response := dtos.AnalyticsResponse{
		TotalPoints:     totalPoints,
		TotalInvestment: totalInvestment,
		BaselineROI:     baselineROI,
	}

	response.RegionAnalytics = make([]dtos.RegionAnalytics, len(regionAnalytics))
	for i, ra := range regionAnalytics {
		response.RegionAnalytics[i] = dtos.RegionAnalytics{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) teamAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teamAnalytics, baselineROI, err := s.analyticsService.GetTeamAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting team analytics: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.AnalyticsResponse{
		BaselineROI: baselineROI,
	}

	response.TeamAnalytics = make([]dtos.TeamAnalytics, len(teamAnalytics))
	for i, ta := range teamAnalytics {
		response.TeamAnalytics[i] = dtos.TeamAnalytics{
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

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) seedVarianceAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	varianceAnalytics, err := s.analyticsService.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting seed variance analytics: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.AnalyticsResponse{}

	response.SeedVarianceAnalytics = make([]dtos.SeedVarianceAnalytics, len(varianceAnalytics))
	for i, sv := range varianceAnalytics {
		response.SeedVarianceAnalytics[i] = dtos.SeedVarianceAnalytics{
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

	writeJSON(w, http.StatusOK, response)
}
