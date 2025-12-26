package calcutta

import (
	"context"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

func (s *Service) GetAllCalcuttas(ctx context.Context) ([]*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetAll(ctx)
}

func (s *Service) GetCalcuttaByID(ctx context.Context, id string) (*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetByID(ctx, id)
}

func (s *Service) CreateCalcutta(ctx context.Context, calcutta *models.Calcutta) error {
	return s.ports.CalcuttaWriter.Create(ctx, calcutta)
}

func (s *Service) CreateRound(ctx context.Context, round *models.CalcuttaRound) error {
	return s.ports.RoundWriter.CreateRound(ctx, round)
}

func (s *Service) UpdateCalcutta(ctx context.Context, calcutta *models.Calcutta) error {
	return s.ports.CalcuttaWriter.Update(ctx, calcutta)
}

func (s *Service) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, error) {
	entries, err := s.ports.EntryReader.GetEntries(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	payouts, err := s.ports.PayoutReader.GetPayouts(ctx, calcuttaID)
	if err != nil {
		return nil, err
	}

	payoutByPosition := map[int]int{}
	for _, p := range payouts {
		payoutByPosition[p.Position] = p.AmountCents
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].TotalPoints == entries[j].TotalPoints {
			return entries[i].Created.After(entries[j].Created)
		}
		return entries[i].TotalPoints > entries[j].TotalPoints
	})

	const epsilon = 0.0001

	position := 1
	for i := 0; i < len(entries); {
		j := i + 1
		for j < len(entries) && math.Abs(entries[j].TotalPoints-entries[i].TotalPoints) < epsilon {
			j++
		}

		groupSize := j - i
		isTied := groupSize > 1

		totalGroupPayout := 0
		for pos := position; pos < position+groupSize; pos++ {
			totalGroupPayout += payoutByPosition[pos]
		}

		base := 0
		remainder := 0
		if groupSize > 0 {
			base = totalGroupPayout / groupSize
			remainder = totalGroupPayout % groupSize
		}

		for k := 0; k < groupSize; k++ {
			e := entries[i+k]
			e.FinishPosition = position
			e.IsTied = isTied
			e.PayoutCents = base
			if remainder > 0 {
				e.PayoutCents++
				remainder--
			}
			e.InTheMoney = e.PayoutCents > 0
		}

		position += groupSize
		i = j
	}

	return entries, nil
}

func (s *Service) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	return s.ports.EntryReader.GetEntryTeams(ctx, entryID)
}

func (s *Service) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	return s.ports.EntryReader.GetEntry(ctx, id)
}

func (s *Service) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	return s.ports.EntryWriter.ReplaceEntryTeams(ctx, entryID, teams)
}

func (s *Service) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfoliosByEntry(ctx, entryID)
}

func (s *Service) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return s.ports.TeamReader.GetTournamentTeam(ctx, id)
}

func (s *Service) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetCalcuttasByTournament(ctx, tournamentID)
}
