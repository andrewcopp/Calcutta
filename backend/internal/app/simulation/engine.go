package simulation

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type mapProbabilityProvider struct {
	probs map[MatchupKey]float64
}

func (p mapProbabilityProvider) Prob(gameID string, team1ID string, team2ID string) float64 {
	if p.probs == nil {
		return 0.5
	}
	if v, ok := p.probs[MatchupKey{GameID: gameID, Team1ID: team1ID, Team2ID: team2ID}]; ok {
		return v
	}
	return 0.5
}

func Simulate(
	bracket *models.BracketStructure,
	probs map[MatchupKey]float64,
	nSims int,
	seed int64,
	opts Options,
) ([]TeamSimulationResult, error) {
	return SimulateWithProvider(bracket, mapProbabilityProvider{probs: probs}, nSims, seed, opts)
}

func SimulateWithProvider(
	bracket *models.BracketStructure,
	provider ProbabilityProvider,
	nSims int,
	seed int64,
	opts Options,
) ([]TeamSimulationResult, error) {
	if bracket == nil {
		return nil, errors.New("bracket must not be nil")
	}
	if len(bracket.Games) == 0 {
		return nil, errors.New("bracket must have games")
	}
	if nSims <= 0 {
		return nil, errors.New("nSims must be positive")
	}

	games, prevByNext := prepareGames(bracket)
	teams, baseByes := collectTeams(games)
	if len(teams) == 0 {
		return nil, errors.New("bracket has no teams")
	}

	workers := opts.Workers
	if workers <= 0 {
		workers = 1
	}

	results := make([]TeamSimulationResult, nSims*len(teams))

	workCh := make(chan int)
	errCh := make(chan error, workers)
	doneCh := make(chan struct{})

	for w := 0; w < workers; w++ {
		go func() {
			for simID := range workCh {
				if err := runOneSimulation(
					simID,
					seed,
					games,
					prevByNext,
					teams,
					baseByes,
					provider,
					results,
				); err != nil {
					errCh <- err
					return
				}
			}
			doneCh <- struct{}{}
		}()
	}

	for simID := 0; simID < nSims; simID++ {
		workCh <- simID
	}
	close(workCh)

	for i := 0; i < workers; i++ {
		select {
		case err := <-errCh:
			return nil, err
		case <-doneCh:
		}
	}

	return results, nil
}

func prepareGames(bracket *models.BracketStructure) ([]*models.BracketGame, map[string]map[int]string) {
	games := make([]*models.BracketGame, 0, len(bracket.Games))
	prevByNext := make(map[string]map[int]string)

	for _, g := range bracket.Games {
		if g == nil {
			continue
		}
		games = append(games, g)
		if g.NextGameID != "" && (g.NextGameSlot == 1 || g.NextGameSlot == 2) {
			slots := prevByNext[g.NextGameID]
			if slots == nil {
				slots = make(map[int]string)
				prevByNext[g.NextGameID] = slots
			}
			slots[g.NextGameSlot] = g.GameID
		}
	}

	sort.Slice(games, func(i, j int) bool {
		gi := games[i]
		gj := games[j]

		ri := gi.Round.Order()
		rj := gj.Round.Order()
		if ri != rj {
			return ri < rj
		}
		if gi.SortOrder != gj.SortOrder {
			return gi.SortOrder < gj.SortOrder
		}
		return gi.GameID < gj.GameID
	})

	return games, prevByNext
}

func collectTeams(games []*models.BracketGame) ([]string, map[string]int) {
	seen := make(map[string]struct{})
	minRound := make(map[string]int)

	for _, g := range games {
		if g == nil {
			continue
		}
		ro := g.Round.Order()
		if g.Team1 != nil && g.Team1.TeamID != "" {
			seen[g.Team1.TeamID] = struct{}{}
			if prev, ok := minRound[g.Team1.TeamID]; !ok || ro < prev {
				minRound[g.Team1.TeamID] = ro
			}
		}
		if g.Team2 != nil && g.Team2.TeamID != "" {
			seen[g.Team2.TeamID] = struct{}{}
			if prev, ok := minRound[g.Team2.TeamID]; !ok || ro < prev {
				minRound[g.Team2.TeamID] = ro
			}
		}
	}

	teams := make([]string, 0, len(seen))
	for tid := range seen {
		teams = append(teams, tid)
	}
	sort.Strings(teams)

	baseByes := make(map[string]int, len(teams))
	firstFourOrder := models.RoundFirstFour.Order()
	for _, tid := range teams {
		ro := minRound[tid]
		if ro != firstFourOrder {
			baseByes[tid] = 1
		} else {
			baseByes[tid] = 0
		}
	}

	return teams, baseByes
}

func runOneSimulation(
	simID int,
	seed int64,
	games []*models.BracketGame,
	prevByNext map[string]map[int]string,
	teams []string,
	baseByes map[string]int,
	provider ProbabilityProvider,
	out []TeamSimulationResult,
) error {
	if simID < 0 {
		return fmt.Errorf("simID must be non-negative")
	}

	rng := rand.New(rand.NewSource(seed + int64(simID)*1_000_003))

	wins := make(map[string]int, len(teams))
	eliminated := make(map[string]bool, len(teams))
	winnersByGame := make(map[string]string, len(games))

	for _, tid := range teams {
		wins[tid] = 0
		eliminated[tid] = false
	}

	for _, g := range games {
		if g == nil || g.GameID == "" {
			continue
		}

		team1 := ""
		team2 := ""
		if g.Team1 != nil {
			team1 = g.Team1.TeamID
		}
		if g.Team2 != nil {
			team2 = g.Team2.TeamID
		}

		if team1 == "" {
			if prev := prevByNext[g.GameID][1]; prev != "" {
				team1 = winnersByGame[prev]
			}
		}
		if team2 == "" {
			if prev := prevByNext[g.GameID][2]; prev != "" {
				team2 = winnersByGame[prev]
			}
		}

		if team1 == "" || team2 == "" {
			continue
		}

		p1 := 0.5
		if provider != nil {
			p1 = provider.Prob(g.GameID, team1, team2)
		}

		roll := rng.Float64()
		winner := team2
		loser := team1
		if roll < p1 {
			winner = team1
			loser = team2
		}

		winnersByGame[g.GameID] = winner
		wins[winner] = wins[winner] + 1
		eliminated[loser] = true
	}

	base := simID * len(teams)
	for i, tid := range teams {
		out[base+i] = TeamSimulationResult{
			SimID:      simID,
			TeamID:     tid,
			Wins:       wins[tid],
			Byes:       baseByes[tid],
			Eliminated: eliminated[tid],
		}
	}

	return nil
}
