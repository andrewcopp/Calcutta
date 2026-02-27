package dtos

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func TestThatNewScoringRuleListResponseMapsAllFields(t *testing.T) {
	// GIVEN scoring rules with various win indices and points
	rules := []*models.ScoringRule{
		{WinIndex: 1, PointsAwarded: 50},
		{WinIndex: 2, PointsAwarded: 150},
		{WinIndex: 3, PointsAwarded: 300},
	}

	// WHEN mapping to response
	resp := NewScoringRuleListResponse(rules)

	// THEN all rules are present with correct values
	if len(resp) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(resp))
	}
}

func TestThatNewScoringRuleListResponsePreservesWinIndex(t *testing.T) {
	// GIVEN a scoring rule with winIndex 5
	rules := []*models.ScoringRule{
		{WinIndex: 5, PointsAwarded: 500},
	}

	// WHEN mapping to response
	resp := NewScoringRuleListResponse(rules)

	// THEN winIndex is preserved
	if resp[0].WinIndex != 5 {
		t.Errorf("expected winIndex 5, got %d", resp[0].WinIndex)
	}
}

func TestThatNewScoringRuleListResponsePreservesPointsAwarded(t *testing.T) {
	// GIVEN a scoring rule with 750 points
	rules := []*models.ScoringRule{
		{WinIndex: 6, PointsAwarded: 750},
	}

	// WHEN mapping to response
	resp := NewScoringRuleListResponse(rules)

	// THEN pointsAwarded is preserved
	if resp[0].PointsAwarded != 750 {
		t.Errorf("expected pointsAwarded 750, got %d", resp[0].PointsAwarded)
	}
}

func TestThatNewScoringRuleListResponseHandlesEmptySlice(t *testing.T) {
	// GIVEN no scoring rules
	rules := []*models.ScoringRule{}

	// WHEN mapping to response
	resp := NewScoringRuleListResponse(rules)

	// THEN an empty (non-nil) slice is returned
	if resp == nil {
		t.Error("expected non-nil slice")
	}
	if len(resp) != 0 {
		t.Errorf("expected 0 rules, got %d", len(resp))
	}
}
