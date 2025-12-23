package main

import (
	"database/sql"
	"testing"
)

func TestChooseShrinkAlphaLOOCVKenPomScore_TieBreakPrefersSmallerAlpha(t *testing.T) {
	// Construct training data where:
	// - All bid shares are identical within seed.
	// - KenPomScore OLS will produce a constant predictor.
	// - Seed baseline mean equals that constant.
	// Therefore every alpha yields identical loss (0), and tie-break should pick smallest alpha.
	train := []trainingBidRow{
		{CalcuttaID: "c1", TeamID: "t1", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c2", TeamID: "t2", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
	}
	alpha := chooseShrinkAlphaLOOCVKenPomScore(train, []float64{0, 0.25, 0.5, 0.75, 1.0})
	if alpha != 0 {
		t.Fatalf("expected alpha=0 due to tie-break, got %v", alpha)
	}
}

func TestChooseShrinkAlphaLOOCVKenPomScore_InsufficientFoldsDefaults(t *testing.T) {
	train := []trainingBidRow{
		{CalcuttaID: "c1", TeamID: "t1", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c1", TeamID: "t2", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 12, Valid: true}, BidShare: 0.2},
	}
	alpha := chooseShrinkAlphaLOOCVKenPomScore(train, []float64{0, 0.5, 1.0})
	if alpha != 0.5 {
		t.Fatalf("expected default alpha=0.5 with <2 folds, got %v", alpha)
	}
}

func TestChooseShrinkAlphaSegmentsLOOCVKenPomScore_TieBreakPrefersMoreShrink(t *testing.T) {
	// Build data such that every alpha-combo yields identical loss (0), so we pick the most conservative combo.
	// Include seeds across segments: 1 (top4), 6 (top8), 12 (other).
	train := []trainingBidRow{
		{CalcuttaID: "c1", TeamID: "t1", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c1", TeamID: "t2", Seed: 6, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c1", TeamID: "t3", Seed: 12, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c2", TeamID: "t4", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c2", TeamID: "t5", Seed: 6, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c2", TeamID: "t6", Seed: 12, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
	}
	a4, a8, ao := chooseShrinkAlphaSegmentsLOOCVKenPomScore(train, []float64{0, 0.5, 1.0})
	if a4 != 0 || a8 != 0 || ao != 0 {
		t.Fatalf("expected (0,0,0) due to conservative tie-break, got (%v,%v,%v)", a4, a8, ao)
	}
}

func TestChooseShrinkAlphaSegmentsLOOCVKenPomScore_InsufficientFoldsDefaults(t *testing.T) {
	train := []trainingBidRow{
		{CalcuttaID: "c1", TeamID: "t1", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1},
		{CalcuttaID: "c1", TeamID: "t2", Seed: 6, KenPomNetRtg: sql.NullFloat64{Float64: 12, Valid: true}, BidShare: 0.2},
		{CalcuttaID: "c1", TeamID: "t3", Seed: 12, KenPomNetRtg: sql.NullFloat64{Float64: 9, Valid: true}, BidShare: 0.05},
	}
	a4, a8, ao := chooseShrinkAlphaSegmentsLOOCVKenPomScore(train, []float64{0, 0.5, 1.0})
	if a4 != 0.5 || a8 != 0.5 || ao != 0.5 {
		t.Fatalf("expected default (0.5,0.5,0.5) with <2 folds, got (%v,%v,%v)", a4, a8, ao)
	}
}

func TestChooseShrinkAlphaSegmentsLOOCVRidgeReturns_InsufficientFoldsDefaults(t *testing.T) {
	train := []trainReturnsRow{
		{CalcuttaID: "c1", TeamID: "t1", Seed: 1, KenPomNetRtg: sql.NullFloat64{Float64: 10, Valid: true}, BidShare: 0.1, SeedBaseline: 0.1, ExpPointsShare: 0.1, ROIFeature: 0.0},
		{CalcuttaID: "c1", TeamID: "t2", Seed: 6, KenPomNetRtg: sql.NullFloat64{Float64: 12, Valid: true}, BidShare: 0.2, SeedBaseline: 0.2, ExpPointsShare: 0.2, ROIFeature: 0.0},
		{CalcuttaID: "c1", TeamID: "t3", Seed: 12, KenPomNetRtg: sql.NullFloat64{Float64: 9, Valid: true}, BidShare: 0.05, SeedBaseline: 0.05, ExpPointsShare: 0.05, ROIFeature: 0.0},
	}
	a4, a8, ao := chooseShrinkAlphaSegmentsLOOCVRidgeReturns(train, []float64{0, 0.5, 1.0})
	if a4 != 0.5 || a8 != 0.5 || ao != 0.5 {
		t.Fatalf("expected default (0.5,0.5,0.5) with <2 folds, got (%v,%v,%v)", a4, a8, ao)
	}
}
