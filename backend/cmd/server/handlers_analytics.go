package main

import (
	"log"
	"net/http"

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
