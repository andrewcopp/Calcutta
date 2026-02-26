package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
)

func (s *Server) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := s.app.Analytics.GetAllAnalytics(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToAnalyticsResponse(result)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) hofBestTeamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestInvestments(ctx, limit)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToBestTeamsResponse(results)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) hofBestInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestInvestmentBids(ctx, limit)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToInvestmentLeaderboardResponse(results)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) hofBestEntriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestEntries(ctx, limit)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToPortfolioLeaderboardResponse(results)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) hofBestCareersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestCareers(ctx, limit)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToCareerLeaderboardResponse(results)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) bestInvestmentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := getLimit(r, 100)

	results, err := s.app.Analytics.GetBestInvestments(ctx, limit)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToBestInvestmentsResponse(results)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) seedInvestmentDistributionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	distribution, err := s.app.Analytics.GetSeedInvestmentDistribution(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.ToSeedInvestmentDistributionResponse(distribution)

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) seedAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	seedAnalytics, totalPoints, totalInvestment, err := s.app.Analytics.GetSeedAnalytics(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	resp := dtos.AnalyticsResponse{
		TotalPoints:     totalPoints,
		TotalInvestment: totalInvestment,
		BaselineROI:     baselineROI,
	}
	resp.SeedAnalytics = make([]dtos.SeedAnalytics, len(seedAnalytics))
	for i, sa := range seedAnalytics {
		resp.SeedAnalytics[i] = dtos.SeedAnalytics{
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

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) regionAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	regionAnalytics, totalPoints, totalInvestment, err := s.app.Analytics.GetRegionAnalytics(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	var baselineROI float64
	if totalInvestment > 0 {
		baselineROI = totalPoints / totalInvestment
	}

	resp := dtos.AnalyticsResponse{
		TotalPoints:     totalPoints,
		TotalInvestment: totalInvestment,
		BaselineROI:     baselineROI,
	}
	resp.RegionAnalytics = make([]dtos.RegionAnalytics, len(regionAnalytics))
	for i, ra := range regionAnalytics {
		resp.RegionAnalytics[i] = dtos.RegionAnalytics{
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

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) teamAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teamAnalytics, baselineROI, err := s.app.Analytics.GetTeamAnalytics(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.AnalyticsResponse{
		BaselineROI: baselineROI,
	}
	resp.TeamAnalytics = make([]dtos.TeamAnalytics, len(teamAnalytics))
	for i, ta := range teamAnalytics {
		resp.TeamAnalytics[i] = dtos.TeamAnalytics{
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

	response.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) seedVarianceAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	varianceAnalytics, err := s.app.Analytics.GetSeedVarianceAnalytics(ctx)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	resp := dtos.AnalyticsResponse{}
	resp.SeedVarianceAnalytics = make([]dtos.SeedVarianceAnalytics, len(varianceAnalytics))
	for i, sv := range varianceAnalytics {
		resp.SeedVarianceAnalytics[i] = dtos.SeedVarianceAnalytics{
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

	response.WriteJSON(w, http.StatusOK, resp)
}
