package httpserver

import (
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
)

func (s *Server) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := s.app.Analytics.GetAllAnalytics(ctx)
	if err != nil {
		log.Printf("Error getting analytics: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToAnalyticsResponse(result)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestTeamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestInvestments(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best teams: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToBestTeamsResponse(results)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestInvestmentBids(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best investments: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToInvestmentLeaderboardResponse(results)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestEntriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestEntries(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best entries: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToEntryLeaderboardResponse(results)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) hofBestCareersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestCareers(ctx, limit)
	if err != nil {
		log.Printf("Error getting hall of fame best careers: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToCareerLeaderboardResponse(results)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) bestInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestInvestments(ctx, limit)
	if err != nil {
		log.Printf("Error getting best investments: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToBestInvestmentsResponse(results)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) seedInvestmentDistributionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	distribution, err := s.app.Analytics.GetSeedInvestmentDistribution(ctx)
	if err != nil {
		log.Printf("Error getting seed investment distribution: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	response := dtos.ToSeedInvestmentDistributionResponse(distribution)

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) seedAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seedAnalytics, totalPoints, totalInvestment, err := s.app.Analytics.GetSeedAnalytics(ctx)
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

	regionAnalytics, totalPoints, totalInvestment, err := s.app.Analytics.GetRegionAnalytics(ctx)
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

	teamAnalytics, baselineROI, err := s.app.Analytics.GetTeamAnalytics(ctx)
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

	varianceAnalytics, err := s.app.Analytics.GetSeedVarianceAnalytics(ctx)
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
