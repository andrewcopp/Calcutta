package prediction

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// CheckpointData holds prediction data for a single tournament checkpoint.
type CheckpointData struct {
	ThroughRound int
	PTVByTeam    map[string]PredictedTeamValue
}

// LoadCheckpointPredictions loads all prediction batches for a tournament,
// deduplicates by through_round (keeping the most recent), and returns
// a slice of checkpoint data.
func (s *Service) LoadCheckpointPredictions(ctx context.Context, tournamentID string) []CheckpointData {
	batches, err := s.ListBatches(ctx, tournamentID)
	if err != nil || len(batches) == 0 {
		return nil
	}

	seen := make(map[int]bool)
	var uniqueBatches []PredictionBatch
	for _, b := range batches {
		if !seen[b.ThroughRound] {
			seen[b.ThroughRound] = true
			uniqueBatches = append(uniqueBatches, b)
		}
	}

	var result []CheckpointData
	for _, b := range uniqueBatches {
		teamValues, err := s.GetTeamValues(ctx, b.ID)
		if err != nil {
			continue
		}
		ptvByTeam := make(map[string]PredictedTeamValue, len(teamValues))
		for _, tv := range teamValues {
			ptvByTeam[tv.TeamID] = tv
		}
		result = append(result, CheckpointData{
			ThroughRound: b.ThroughRound,
			PTVByTeam:    ptvByTeam,
		})
	}

	return result
}

// LatestCheckpoint returns the checkpoint with the highest ThroughRound.
func LatestCheckpoint(checkpoints []CheckpointData) *CheckpointData {
	if len(checkpoints) == 0 {
		return nil
	}
	best := &checkpoints[0]
	for i := 1; i < len(checkpoints); i++ {
		if checkpoints[i].ThroughRound > best.ThroughRound {
			best = &checkpoints[i]
		}
	}
	return best
}

// BestCheckpointForCap returns the checkpoint with the highest ThroughRound <= cap.
// Returns nil if no checkpoint qualifies.
func BestCheckpointForCap(checkpoints []CheckpointData, cap int) *CheckpointData {
	var best *CheckpointData
	for i := range checkpoints {
		if checkpoints[i].ThroughRound <= cap {
			if best == nil || checkpoints[i].ThroughRound > best.ThroughRound {
				best = &checkpoints[i]
			}
		}
	}
	return best
}

// PortfolioTeamInput is the minimal data needed from a portfolio team for projection math.
type PortfolioTeamInput struct {
	PortfolioID         string
	TeamID              string
	OwnershipPercentage float64
}

// TournamentTeamInput is the minimal data needed from a tournament team for projection math.
type TournamentTeamInput struct {
	ID           string
	Wins         int
	Byes         int
	IsEliminated bool
}

// EntryProjections holds projected EV and Favorites for each entry.
type EntryProjections struct {
	EV        map[string]float64
	Favorites map[string]float64
}

// ComputeEntryProjections computes projected EV and Favorites for each entry
// using the latest checkpoint. Returns nil if no checkpoints are available.
func ComputeEntryProjections(
	checkpoints []CheckpointData,
	rules []scoring.Rule,
	portfolioToEntry map[string]string,
	portfolioTeams []PortfolioTeamInput,
	tournamentTeams []TournamentTeamInput,
) *EntryProjections {
	cp := LatestCheckpoint(checkpoints)
	if cp == nil {
		return nil
	}
	return computeProjectionsForCheckpoint(cp, rules, portfolioToEntry, portfolioTeams, tournamentTeams)
}

// ProgressAtRound returns the effective progress (wins + byes) capped at the given round.
func ProgressAtRound(wins, byes, round int) int {
	progress := wins + byes
	if progress > round {
		return round
	}
	return progress
}

// snapshotTeamsAtRound returns a copy of teams with progress capped at roundCap.
// Byes are folded into wins, and teams eliminated beyond the cap are treated as alive.
func snapshotTeamsAtRound(teams []TournamentTeamInput, roundCap int) []TournamentTeamInput {
	out := make([]TournamentTeamInput, len(teams))
	for i, tt := range teams {
		progress := tt.Wins + tt.Byes
		out[i] = TournamentTeamInput{
			ID:           tt.ID,
			Wins:         ProgressAtRound(tt.Wins, tt.Byes, roundCap),
			Byes:         0,
			IsEliminated: tt.IsEliminated && progress <= roundCap,
		}
	}
	return out
}

// ComputeRoundProjections computes projected EV and Favorites for each entry
// at a specific round cap using the best matching checkpoint.
// Returns nil if no suitable checkpoint exists.
func ComputeRoundProjections(
	checkpoints []CheckpointData,
	rules []scoring.Rule,
	portfolioToEntry map[string]string,
	portfolioTeams []PortfolioTeamInput,
	tournamentTeams []TournamentTeamInput,
	roundCap int,
) *EntryProjections {
	cp := BestCheckpointForCap(checkpoints, roundCap)
	if cp == nil {
		return nil
	}
	return computeProjectionsForCheckpoint(cp, rules, portfolioToEntry, portfolioTeams, snapshotTeamsAtRound(tournamentTeams, roundCap))
}

func computeProjectionsForCheckpoint(
	cp *CheckpointData,
	rules []scoring.Rule,
	portfolioToEntry map[string]string,
	portfolioTeams []PortfolioTeamInput,
	tournamentTeams []TournamentTeamInput,
) *EntryProjections {
	ttByID := make(map[string]TournamentTeamInput, len(tournamentTeams))
	for _, tt := range tournamentTeams {
		ttByID[tt.ID] = tt
	}

	ev := make(map[string]float64)
	fav := make(map[string]float64)
	for _, pt := range portfolioTeams {
		ptv, ok := cp.PTVByTeam[pt.TeamID]
		if !ok {
			continue
		}
		tt, ok := ttByID[pt.TeamID]
		if !ok {
			continue
		}
		entryID := portfolioToEntry[pt.PortfolioID]
		if entryID == "" {
			continue
		}
		tp := TeamProgress{
			Wins:         tt.Wins,
			Byes:         tt.Byes,
			IsEliminated: tt.IsEliminated,
		}
		ev[entryID] += pt.OwnershipPercentage * ProjectedTeamEV(ptv, rules, tp, cp.ThroughRound)
		fav[entryID] += pt.OwnershipPercentage * ptv.FavoritesTotalPoints
	}

	return &EntryProjections{EV: ev, Favorites: fav}
}

// BuildPortfolioToEntry builds a portfolio ID → entry ID lookup map.
func BuildPortfolioToEntry(portfolios []*models.CalcuttaPortfolio) map[string]string {
	m := make(map[string]string, len(portfolios))
	for _, p := range portfolios {
		m[p.ID] = p.EntryID
	}
	return m
}

// ToPortfolioTeamInputs converts domain portfolio teams to the minimal input type.
func ToPortfolioTeamInputs(pts []*models.CalcuttaPortfolioTeam) []PortfolioTeamInput {
	out := make([]PortfolioTeamInput, len(pts))
	for i, pt := range pts {
		out[i] = PortfolioTeamInput{
			PortfolioID:         pt.PortfolioID,
			TeamID:              pt.TeamID,
			OwnershipPercentage: pt.OwnershipPercentage,
		}
	}
	return out
}

// ToTournamentTeamInputs converts domain tournament teams to the minimal input type.
func ToTournamentTeamInputs(tts []*models.TournamentTeam) []TournamentTeamInput {
	out := make([]TournamentTeamInput, len(tts))
	for i, tt := range tts {
		out[i] = TournamentTeamInput{
			ID:           tt.ID,
			Wins:         tt.Wins,
			Byes:         tt.Byes,
			IsEliminated: tt.IsEliminated,
		}
	}
	return out
}
