package main

import "errors"

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
