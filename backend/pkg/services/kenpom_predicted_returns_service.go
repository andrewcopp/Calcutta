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

func kenPomAdjEM(t *models.TournamentTeam) (float64, bool) {
	if t == nil || t.KenPom == nil || t.KenPom.ORtg == nil || t.KenPom.DRtg == nil {
		return 0, false
	}
	return *t.KenPom.ORtg - *t.KenPom.DRtg, true
}

func kenPomExpectedMargin(teamA *models.TournamentTeam, teamB *models.TournamentTeam) (float64, bool) {
	if teamA == nil || teamB == nil {
		return 0, false
	}
	if teamA.KenPom == nil || teamB.KenPom == nil {
		return 0, false
	}
	if teamA.KenPom.AdjT == nil || teamB.KenPom.AdjT == nil {
		return 0, false
	}
	adjEMA, okA := kenPomAdjEM(teamA)
	adjEMB, okB := kenPomAdjEM(teamB)
	if !okA || !okB {
		return 0, false
	}

	marginPer100 := adjEMA - adjEMB
	poss := (*teamA.KenPom.AdjT + *teamB.KenPom.AdjT) / 2.0
	return marginPer100 * (poss / 100.0), true
}

func kenPomWinProbFromMargin(margin float64, sigma float64) float64 {
	if sigma <= 0 {
		sigma = 11.0
	}
	z := margin / sigma
	p := 0.5 * (1.0 + math.Erf(z/math.Sqrt2))
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return p
}

func kenPomWinProb(teamA *models.TournamentTeam, teamB *models.TournamentTeam, sigma float64) (float64, error) {
	margin, ok := kenPomExpectedMargin(teamA, teamB)
	if !ok {
		return 0, fmt.Errorf("missing/incomplete kenpom stats for matchup")
	}
	return kenPomWinProbFromMargin(margin, sigma), nil
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
	tournament, err := s.bracketService.tournamentRepo.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}
	if tournament == nil {
		return nil, fmt.Errorf("tournament not found")
	}

	teams, err := s.bracketService.tournamentRepo.GetTeams(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	finalFour := &models.FinalFourConfig{
		TopLeftRegion:     tournament.FinalFourTopLeft,
		BottomLeftRegion:  tournament.FinalFourBottomLeft,
		TopRightRegion:    tournament.FinalFourTopRight,
		BottomRightRegion: tournament.FinalFourBottomRight,
	}
	if finalFour.TopLeftRegion == "" {
		finalFour.TopLeftRegion = "East"
	}
	if finalFour.BottomLeftRegion == "" {
		finalFour.BottomLeftRegion = "West"
	}
	if finalFour.TopRightRegion == "" {
		finalFour.TopRightRegion = "South"
	}
	if finalFour.BottomRightRegion == "" {
		finalFour.BottomRightRegion = "Midwest"
	}

	bracket, err := s.bracketService.builder.BuildBracket(tournamentID, teams, finalFour)
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
		p, err := kenPomWinProb(a, b, s.sigma)
		if err != nil {
			return 0, fmt.Errorf("%w: %s vs %s", err, teamAID, teamBID)
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
