package services

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type KenPomPredictedReturn struct {
	TeamID         string
	SchoolID       string
	SchoolName     string
	Seed           int
	Region         string
	ExpectedPoints float64
}

type KenPomPredictedReturnsService struct {
	bracketService *BracketService
	sigma          float64
}

func NewKenPomPredictedReturnsService(bracketService *BracketService) *KenPomPredictedReturnsService {
	return &KenPomPredictedReturnsService{
		bracketService: bracketService,
		sigma:          11.0,
	}
}

func (s *KenPomPredictedReturnsService) WithSigma(sigma float64) *KenPomPredictedReturnsService {
	if sigma > 0 {
		s.sigma = sigma
	}
	return s
}

func (s *KenPomPredictedReturnsService) GetPredictedReturnsPreTournament(ctx context.Context, tournamentID string) ([]KenPomPredictedReturn, error) {
	bracket, err := s.bracketService.GetBracket(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	teams, err := s.bracketService.tournamentRepo.GetTeams(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	teamsByID := make(map[string]*models.TournamentTeam, len(teams))
	for _, t := range teams {
		if t != nil {
			teamsByID[t.ID] = t
		}
	}

	sources := make(map[string]struct{ slot1, slot2 string }, len(bracket.Games))
	for _, g := range bracket.Games {
		if g == nil {
			continue
		}
		if g.NextGameID == "" {
			continue
		}
		s := sources[g.NextGameID]
		if g.NextGameSlot == 1 {
			s.slot1 = g.GameID
		} else if g.NextGameSlot == 2 {
			s.slot2 = g.GameID
		}
		sources[g.NextGameID] = s
	}

	roundOrder := []models.BracketRound{
		models.RoundFirstFour,
		models.RoundOf64,
		models.RoundOf32,
		models.RoundSweet16,
		models.RoundElite8,
		models.RoundFinalFour,
		models.RoundChampionship,
	}

	pointsByRound := map[models.BracketRound]float64{
		models.RoundFirstFour:    0,
		models.RoundOf64:         50,
		models.RoundOf32:         100,
		models.RoundSweet16:      150,
		models.RoundElite8:       200,
		models.RoundFinalFour:    250,
		models.RoundChampionship: 300,
	}

	firstFourLosers := make(map[string]bool)
	for _, g := range bracket.Games {
		if g == nil || g.Round != models.RoundFirstFour {
			continue
		}
		if g.Team1 == nil || g.Team2 == nil {
			return nil, fmt.Errorf("first four game %s missing participants", g.GameID)
		}
		if g.Winner == nil {
			return nil, fmt.Errorf("first four game %s has no winner selected", g.GameID)
		}
		if g.Winner.TeamID == g.Team1.TeamID {
			firstFourLosers[g.Team2.TeamID] = true
		} else if g.Winner.TeamID == g.Team2.TeamID {
			firstFourLosers[g.Team1.TeamID] = true
		} else {
			return nil, fmt.Errorf("first four game %s winner is not a participant", g.GameID)
		}
	}

	type dist map[string]float64

	winDistByGame := make(map[string]dist, len(bracket.Games))
	evByTeam := make(map[string]float64, len(teamsByID))

	getSlotDist := func(game *models.BracketGame, slot int) (dist, error) {
		if slot != 1 && slot != 2 {
			return nil, fmt.Errorf("invalid slot %d", slot)
		}
		s := sources[game.GameID]
		if slot == 1 {
			if s.slot1 != "" {
				d, ok := winDistByGame[s.slot1]
				if !ok {
					return nil, fmt.Errorf("missing win distribution for game %s", s.slot1)
				}
				return d, nil
			}
			if game.Team1 == nil {
				return nil, fmt.Errorf("missing team1 for game %s", game.GameID)
			}
			return dist{game.Team1.TeamID: 1}, nil
		}

		if s.slot2 != "" {
			d, ok := winDistByGame[s.slot2]
			if !ok {
				return nil, fmt.Errorf("missing win distribution for game %s", s.slot2)
			}
			return d, nil
		}
		if game.Team2 == nil {
			return nil, fmt.Errorf("missing team2 for game %s", game.GameID)
		}
		return dist{game.Team2.TeamID: 1}, nil
	}

	winProb := func(teamAID, teamBID string) (float64, error) {
		a := teamsByID[teamAID]
		b := teamsByID[teamBID]
		if a == nil || b == nil {
			return 0, fmt.Errorf("missing team data for matchup %s vs %s", teamAID, teamBID)
		}
		if a.KenPom == nil || b.KenPom == nil {
			return 0, fmt.Errorf("missing kenpom stats for matchup %s vs %s", teamAID, teamBID)
		}
		if a.KenPom.ORtg == nil || a.KenPom.DRtg == nil || a.KenPom.AdjT == nil {
			return 0, fmt.Errorf("incomplete kenpom stats for team %s", teamAID)
		}
		if b.KenPom.ORtg == nil || b.KenPom.DRtg == nil || b.KenPom.AdjT == nil {
			return 0, fmt.Errorf("incomplete kenpom stats for team %s", teamBID)
		}

		m100 := (*a.KenPom.ORtg - *b.KenPom.DRtg) - (*b.KenPom.ORtg - *a.KenPom.DRtg)
		poss := (*a.KenPom.AdjT + *b.KenPom.AdjT) / 2.0
		m := m100 * (poss / 100.0)

		z := m / s.sigma
		p := 0.5 * (1.0 + math.Erf(z/math.Sqrt2))
		if p < 0 {
			p = 0
		}
		if p > 1 {
			p = 1
		}
		return p, nil
	}

	for _, round := range roundOrder {
		games := make([]*models.BracketGame, 0)
		for _, g := range bracket.Games {
			if g == nil || g.Round != round {
				continue
			}
			games = append(games, g)
		}
		sort.Slice(games, func(i, j int) bool { return games[i].SortOrder < games[j].SortOrder })

		for _, g := range games {
			if g.Round == models.RoundFirstFour && g.Winner != nil {
				winDistByGame[g.GameID] = dist{g.Winner.TeamID: 1}
				continue
			}

			left, err := getSlotDist(g, 1)
			if err != nil {
				return nil, err
			}
			right, err := getSlotDist(g, 2)
			if err != nil {
				return nil, err
			}

			win := dist{}
			for t1, p1 := range left {
				for t2, p2 := range right {
					p, err := winProb(t1, t2)
					if err != nil {
						return nil, err
					}
					joint := p1 * p2
					win[t1] += joint * p
					win[t2] += joint * (1.0 - p)
				}
			}

			winDistByGame[g.GameID] = win
		}
	}

	for _, g := range bracket.Games {
		if g == nil {
			continue
		}
		d := winDistByGame[g.GameID]
		if d == nil {
			continue
		}
		pts, ok := pointsByRound[g.Round]
		if !ok {
			continue
		}
		if pts == 0 {
			continue
		}
		for teamID, p := range d {
			evByTeam[teamID] += p * pts
		}
	}

	out := make([]KenPomPredictedReturn, 0, len(teamsByID))
	for teamID, team := range teamsByID {
		if team == nil {
			continue
		}
		if firstFourLosers[teamID] {
			continue
		}
		name := ""
		if team.School != nil {
			name = team.School.Name
		}
		out = append(out, KenPomPredictedReturn{
			TeamID:         teamID,
			SchoolID:       team.SchoolID,
			SchoolName:     name,
			Seed:           team.Seed,
			Region:         team.Region,
			ExpectedPoints: evByTeam[teamID],
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ExpectedPoints == out[j].ExpectedPoints {
			if out[i].Seed == out[j].Seed {
				return out[i].SchoolName < out[j].SchoolName
			}
			return out[i].Seed < out[j].Seed
		}
		return out[i].ExpectedPoints > out[j].ExpectedPoints
	})

	return out, nil
}
