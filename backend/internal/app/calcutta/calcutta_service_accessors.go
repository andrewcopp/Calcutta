package calcutta

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

func (s *Service) GetAllCalcuttas(ctx context.Context) ([]*models.Calcutta, error) {
	return s.ports.Calcuttas.GetAll(ctx)
}

func (s *Service) GetCalcuttaByID(ctx context.Context, id string) (*models.Calcutta, error) {
	return s.ports.Calcuttas.GetByID(ctx, id)
}

func (s *Service) UpdateCalcutta(ctx context.Context, calcutta *models.Calcutta) error {
	return s.ports.Calcuttas.Update(ctx, calcutta)
}

func (s *Service) GetEntries(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntry, []*models.EntryStanding, error) {
	entries, pointsByEntry, err := s.ports.Entries.GetEntries(ctx, calcuttaID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting entries: %w", err)
	}

	payouts, err := s.ports.Payouts.GetPayouts(ctx, calcuttaID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting payouts: %w", err)
	}

	standings := ComputeStandings(entries, pointsByEntry, payouts)
	return entries, standings, nil
}

// ComputeStandings computes finish positions and payouts from entries, their points, and payout rules.
// Returns standings sorted by points descending. Does not mutate entries.
func ComputeStandings(
	entries []*models.CalcuttaEntry,
	pointsByEntry map[string]float64,
	payouts []*models.CalcuttaPayout,
) []*models.EntryStanding {
	if entries == nil {
		return nil
	}

	type entryWithPoints struct {
		entry  *models.CalcuttaEntry
		points float64
	}

	sorted := make([]entryWithPoints, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		sorted = append(sorted, entryWithPoints{entry: e, points: pointsByEntry[e.ID]})
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].points == sorted[j].points {
			return sorted[i].entry.CreatedAt.After(sorted[j].entry.CreatedAt)
		}
		return sorted[i].points > sorted[j].points
	})

	payoutByPosition := map[int]int{}
	for _, p := range payouts {
		if p == nil {
			continue
		}
		payoutByPosition[p.Position] = p.AmountCents
	}

	const epsilon = 0.0001
	standings := make([]*models.EntryStanding, len(sorted))

	position := 1
	for i := 0; i < len(sorted); {
		j := i + 1
		for j < len(sorted) {
			if math.Abs(sorted[j].points-sorted[i].points) >= epsilon {
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
			payoutCents := base
			if remainder > 0 {
				payoutCents++
				remainder--
			}
			standings[i+k] = &models.EntryStanding{
				EntryID:        sorted[i+k].entry.ID,
				TotalPoints:    sorted[i+k].points,
				FinishPosition: position,
				IsTied:         isTied,
				PayoutCents:    payoutCents,
				InTheMoney:     payoutCents > 0,
			}
		}

		position += groupSize
		i = j
	}

	return standings
}

func (s *Service) GetEntryTeams(ctx context.Context, entryID string) ([]*models.CalcuttaEntryTeam, error) {
	return s.ports.Entries.GetEntryTeams(ctx, entryID)
}

func (s *Service) GetEntryTeamsByEntryIDs(ctx context.Context, entryIDs []string) (map[string][]*models.CalcuttaEntryTeam, error) {
	return s.ports.Entries.GetEntryTeamsByEntryIDs(ctx, entryIDs)
}

func (s *Service) GetEntry(ctx context.Context, id string) (*models.CalcuttaEntry, error) {
	return s.ports.Entries.GetEntry(ctx, id)
}

func (s *Service) CreateEntry(ctx context.Context, entry *models.CalcuttaEntry) error {
	return s.ports.Entries.CreateEntry(ctx, entry)
}

func (s *Service) ReplaceEntryTeams(ctx context.Context, entryID string, teams []*models.CalcuttaEntryTeam) error {
	return s.ports.Entries.ReplaceEntryTeams(ctx, entryID, teams)
}

func (s *Service) GetPortfoliosByEntry(ctx context.Context, entryID string) ([]*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfoliosByEntry(ctx, entryID)
}

func (s *Service) GetPortfoliosByEntryIDs(ctx context.Context, entryIDs []string) (map[string][]*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfoliosByEntryIDs(ctx, entryIDs)
}

func (s *Service) GetPortfolioTeamsByPortfolioIDs(ctx context.Context, portfolioIDs []string) (map[string][]*models.CalcuttaPortfolioTeam, error) {
	return s.ports.PortfolioReader.GetPortfolioTeamsByPortfolioIDs(ctx, portfolioIDs)
}

func (s *Service) GetPortfolio(ctx context.Context, id string) (*models.CalcuttaPortfolio, error) {
	return s.ports.PortfolioReader.GetPortfolio(ctx, id)
}

func (s *Service) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return s.ports.TeamReader.GetTournamentTeam(ctx, id)
}

func (s *Service) GetCalcuttasByUser(ctx context.Context, userID string) ([]*models.Calcutta, error) {
	return s.ports.Calcuttas.GetByUserID(ctx, userID)
}

func (s *Service) GetRounds(ctx context.Context, calcuttaID string) ([]*models.CalcuttaRound, error) {
	return s.ports.Rounds.GetRounds(ctx, calcuttaID)
}

func (s *Service) GetCalcuttasByTournament(ctx context.Context, tournamentID string) ([]*models.Calcutta, error) {
	return s.ports.Calcuttas.GetCalcuttasByTournament(ctx, tournamentID)
}

func (s *Service) GetDistinctUserIDsByCalcutta(ctx context.Context, calcuttaID string) ([]string, error) {
	return s.ports.Entries.GetDistinctUserIDsByCalcutta(ctx, calcuttaID)
}

func (s *Service) GetPayouts(ctx context.Context, calcuttaID string) ([]*models.CalcuttaPayout, error) {
	return s.ports.Payouts.GetPayouts(ctx, calcuttaID)
}

func (s *Service) ReplacePayouts(ctx context.Context, calcuttaID string, payouts []*models.CalcuttaPayout) error {
	return s.ports.Payouts.ReplacePayouts(ctx, calcuttaID, payouts)
}
