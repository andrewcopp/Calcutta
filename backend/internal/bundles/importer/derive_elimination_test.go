package importer

import (
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
)

func TestThatDeriveIsEliminatedMarksAllButChampionInCompletedTournament(t *testing.T) {
	// GIVEN a completed tournament where one team has 6 wins (champion)
	teams := []bundles.TeamRecord{
		{SchoolSlug: "champion", Wins: 6},
		{SchoolSlug: "finalist", Wins: 5},
		{SchoolSlug: "round-of-32", Wins: 1},
		{SchoolSlug: "first-round-exit", Wins: 0},
	}

	// WHEN deriving elimination status
	deriveIsEliminated(teams)

	// THEN only the champion is not eliminated
	for _, team := range teams {
		if team.SchoolSlug == "champion" && team.IsEliminated {
			t.Errorf("champion should not be eliminated")
		}
		if team.SchoolSlug != "champion" && !team.IsEliminated {
			t.Errorf("team %s should be eliminated", team.SchoolSlug)
		}
	}
}

func TestThatDeriveIsEliminatedLeavesAllAliveWhenTournamentNotStarted(t *testing.T) {
	// GIVEN a tournament that hasn't started (all teams have 0 wins)
	teams := []bundles.TeamRecord{
		{SchoolSlug: "team-a", Wins: 0},
		{SchoolSlug: "team-b", Wins: 0},
		{SchoolSlug: "team-c", Wins: 0},
	}

	// WHEN deriving elimination status
	deriveIsEliminated(teams)

	// THEN no team is eliminated
	for _, team := range teams {
		if team.IsEliminated {
			t.Errorf("team %s should not be eliminated in unstarted tournament", team.SchoolSlug)
		}
	}
}

func TestThatDeriveIsEliminatedHandlesTeamsWithByes(t *testing.T) {
	// GIVEN teams where progress = wins + byes
	teams := []bundles.TeamRecord{
		{SchoolSlug: "bye-team", Wins: 2, Byes: 1},   // progress = 3
		{SchoolSlug: "no-bye-team", Wins: 3, Byes: 0}, // progress = 3
		{SchoolSlug: "eliminated", Wins: 1, Byes: 0},  // progress = 1
	}

	// WHEN deriving elimination status
	deriveIsEliminated(teams)

	// THEN teams with equal progress are both alive
	if teams[0].IsEliminated {
		t.Error("bye-team with progress 3 should not be eliminated")
	}
	if teams[1].IsEliminated {
		t.Error("no-bye-team with progress 3 should not be eliminated")
	}
	if !teams[2].IsEliminated {
		t.Error("eliminated team with progress 1 should be eliminated")
	}
}

func TestThatDeriveIsEliminatedPreservesExistingEliminationFlags(t *testing.T) {
	// GIVEN teams where some are already marked eliminated
	teams := []bundles.TeamRecord{
		{SchoolSlug: "champion", Wins: 6, IsEliminated: false},
		{SchoolSlug: "already-eliminated", Wins: 2, IsEliminated: true},
		{SchoolSlug: "not-yet-marked", Wins: 1, IsEliminated: false},
	}

	// WHEN deriving elimination status
	deriveIsEliminated(teams)

	// THEN previously eliminated teams stay eliminated (idempotent)
	if !teams[1].IsEliminated {
		t.Error("already-eliminated team should still be eliminated")
	}
}
