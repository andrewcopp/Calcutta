package scoring

import "testing"

func TestThatPointsForProgressReturnsZeroWhenRulesEmpty(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule(nil)
	GIVENWins := 3
	GIVENByes := 0

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, GIVENWins, GIVENByes)

	// THEN
	if WHENPoints != 0 {
		t.Fatalf("expected 0, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressReturnsZeroWhenProgressNonPositive(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{{WinIndex: 1, PointsAwarded: 10}}

	// WHEN
	WHENPoints0 := PointsForProgress(GIVENRules, 0, 0)
	WHENPointsNeg := PointsForProgress(GIVENRules, -1, 0)

	// THEN
	if WHENPoints0 != 0 {
		t.Fatalf("expected 0, got %d", WHENPoints0)
	}
	if WHENPointsNeg != 0 {
		t.Fatalf("expected 0, got %d", WHENPointsNeg)
	}
}

func TestThatPointsForProgressSumsRulesUpToProgress(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{
		{WinIndex: 1, PointsAwarded: 0},
		{WinIndex: 2, PointsAwarded: 50},
		{WinIndex: 3, PointsAwarded: 100},
		{WinIndex: 4, PointsAwarded: 150},
	}
	GIVENWins := 3
	GIVENByes := 0

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, GIVENWins, GIVENByes)

	// THEN
	if WHENPoints != 150 {
		t.Fatalf("expected 150, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressCountsByesAsProgress(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{
		{WinIndex: 1, PointsAwarded: 0},
		{WinIndex: 2, PointsAwarded: 50},
		{WinIndex: 3, PointsAwarded: 100},
	}
	GIVENWins := 2
	GIVENByes := 1

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, GIVENWins, GIVENByes)

	// THEN
	if WHENPoints != 150 {
		t.Fatalf("expected 150, got %d", WHENPoints)
	}
}

func TestThatPointsForProgressIsOrderIndependent(t *testing.T) {
	// GIVEN
	GIVENRules := []Rule{
		{WinIndex: 3, PointsAwarded: 100},
		{WinIndex: 1, PointsAwarded: 0},
		{WinIndex: 2, PointsAwarded: 50},
	}

	// WHEN
	WHENPoints := PointsForProgress(GIVENRules, 3, 0)

	// THEN
	if WHENPoints != 150 {
		t.Fatalf("expected 150, got %d", WHENPoints)
	}
}
