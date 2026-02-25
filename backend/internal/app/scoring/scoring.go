package scoring

type Rule struct {
	WinIndex      int
	PointsAwarded int
}

// TournamentTotal computes the deterministic point total for a bracket.
// gamesPerRound[i] is the number of games in round i+1 (0-indexed).
// For each round, total += gamesPerRound[i] * incrementalPoints[i+1].
// Rules beyond the number of rounds are ignored.
func TournamentTotal(rules []Rule, gamesPerRound []int) int {
	total := 0
	for i, games := range gamesPerRound {
		round := i + 1
		inc := PointsForProgress(rules, round, 0) - PointsForProgress(rules, round-1, 0)
		total += games * inc
	}
	return total
}

// GamesPerRoundForBracket returns games per round for a standard
// single-elimination bracket (e.g., 4 teams -> [2,1], 64 -> [32,16,8,4,2,1]).
func GamesPerRoundForBracket(numTeams int) []int {
	if numTeams < 2 {
		return nil
	}
	var result []int
	remaining := numTeams
	for remaining > 1 {
		games := remaining / 2
		result = append(result, games)
		remaining = games
	}
	return result
}

// NCAAgamesPerRound returns the games per round for the 128-team symmetric bracket model:
// 64 R128 (60 BYE + 4 First Four), 32 R64, 16 R32, 8 S16, 4 E8, 2 FF, 1 Championship.
func NCAAgamesPerRound() []int {
	return []int{64, 32, 16, 8, 4, 2, 1}
}

func PointsForProgress(rules []Rule, wins int, byes int) int {
	p := wins + byes
	if p <= 0 {
		return 0
	}
	pts := 0
	for _, r := range rules {
		if r.WinIndex <= p {
			pts += r.PointsAwarded
		}
	}
	return pts
}
