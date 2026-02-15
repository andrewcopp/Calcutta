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

	sorted, results := ComputeEntryPlacementsAndPayouts(entries, payouts)
	if len(sorted) == len(entries) {
		copy(entries, sorted)
		sorted = entries
	}
	for _, e := range sorted {
		if e == nil {
			continue
		}
		res, ok := results[e.ID]
		if !ok {
			continue
		}
		e.FinishPosition = res.FinishPosition
		e.IsTied = res.IsTied
		e.PayoutCents = res.PayoutCents
		e.InTheMoney = res.InTheMoney
	}
	return sorted, nil
}

type EntryPayoutResult struct {
	FinishPosition int
	IsTied         bool
	PayoutCents    int
	InTheMoney     bool
}

func ComputeEntryPlacementsAndPayouts(entries []*models.CalcuttaEntry, payouts []*models.CalcuttaPayout) ([]*models.CalcuttaEntry, map[string]EntryPayoutResult) {
	if entries == nil {
		return nil, nil
	}

	out := make([]*models.CalcuttaEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, e)
	}

	results := make(map[string]EntryPayoutResult, len(entries))

	payoutByPosition := map[int]int{}
	for _, p := range payouts {
		if p == nil {
			continue
		}
		payoutByPosition[p.Position] = p.AmountCents
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i] == nil {
			return false
		}
		if out[j] == nil {
			return true
		}
		if out[i].TotalPoints == out[j].TotalPoints {
			return out[i].Created.After(out[j].Created)
		}
		return out[i].TotalPoints > out[j].TotalPoints
	})

	const epsilon = 0.0001

	position := 1
	for i := 0; i < len(out); {
		if out[i] == nil {
			i++
			continue
		}

		j := i + 1
		for j < len(out) {
			if out[j] == nil {
				break
			}
			if math.Abs(out[j].TotalPoints-out[i].TotalPoints) >= epsilon {
				break
			}
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
			e := out[i+k]
			if e == nil {
				continue
			}
			payoutCents := base
			if remainder > 0 {
				payoutCents++
				remainder--
			}
			results[e.ID] = EntryPayoutResult{
				FinishPosition: position,
				IsTied:         isTied,
				PayoutCents:    payoutCents,
				InTheMoney:     payoutCents > 0,
			}
		}

		position += groupSize
		i = j
	}

	return out, results
}

func (s *Service) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	return s.ports.EntryReader.GetEntryTeams(ctx, entryID)
}

func (s *Service) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	return s.ports.EntryReader.GetEntry(ctx, id)
}

func (s *Service) CreateEntry(ctx context.Context, entry *models.CalcuttaEntry) error {
	return s.ports.EntryWriter.CreateEntry(ctx, entry)
}

func (s *Service) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	return s.ports.EntryWriter.ReplaceEntryTeams(ctx, entryID, teams)
}

func (s *Service) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfoliosByEntry(ctx, entryID)
}

func (s *Service) GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfolio(ctx, id)
}

func (s *Service) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return s.ports.TeamReader.GetTournamentTeam(ctx, id)
}

func (s *Service) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	return s.ports.CalcuttaReader.GetCalcuttasByTournament(ctx, tournamentID)
}
