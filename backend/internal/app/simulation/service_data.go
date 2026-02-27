package simulation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/app/winprob"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// KenPomProvider implements ProbabilityProvider using KenPom net ratings
// and a logistic model to estimate win probabilities.
type KenPomProvider struct {
	Spec        *winprob.Model
	NetByTeamID map[string]float64
	Overrides   map[MatchupKey]float64
}

// NewKenPomProvider creates a KenPomProvider from a spec, net ratings by team
// ID, and optional per-matchup probability overrides.
func NewKenPomProvider(spec *winprob.Model, netByTeamID map[string]float64, overrides map[MatchupKey]float64) *KenPomProvider {
	return &KenPomProvider{
		Spec:        spec,
		NetByTeamID: netByTeamID,
		Overrides:   overrides,
	}
}

func (p KenPomProvider) Prob(gameID string, team1ID string, team2ID string) float64 {
	if p.Overrides != nil {
		if v, ok := p.Overrides[MatchupKey{GameID: gameID, Team1ID: team1ID, Team2ID: team2ID}]; ok {
			return v
		}
	}
	if p.Spec == nil {
		return 0.5
	}
	n1, ok1 := p.NetByTeamID[team1ID]
	n2, ok2 := p.NetByTeamID[team2ID]
	if !ok1 || !ok2 {
		return 0.5
	}
	return p.Spec.WinProb(n1, n2)
}

func (s *Service) loadKenPomNetByTeamID(ctx context.Context, coreTournamentID string) (map[string]float64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, ks.net_rtg
		FROM core.teams t
		LEFT JOIN core.team_kenpom_stats ks
			ON ks.team_id = t.id
			AND ks.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
	`, coreTournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying kenpom ratings: %w", err)
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var teamID string
		var net *float64
		if err := rows.Scan(&teamID, &net); err != nil {
			return nil, fmt.Errorf("scanning kenpom rating: %w", err)
		}
		if net != nil {
			out[teamID] = *net
		}
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterating kenpom ratings: %w", rows.Err())
	}
	return out, nil
}

func (s *Service) loadTeams(ctx context.Context, coreTournamentID string) ([]*models.TournamentTeam, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id,
			t.seed,
			t.region,
			s.name
		FROM core.teams t
		JOIN core.schools s
			ON s.id = t.school_id
			AND s.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
		ORDER BY t.seed ASC, s.name ASC
	`, coreTournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying teams: %w", err)
	}
	defer rows.Close()

	out := make([]*models.TournamentTeam, 0)
	for rows.Next() {
		var id string
		var seed *int
		var region *string
		var schoolName string
		if err := rows.Scan(&id, &seed, &region, &schoolName); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}

		seedVal := 0
		if seed != nil {
			seedVal = *seed
		}
		regionVal := ""
		if region != nil {
			regionVal = *region
		}

		out = append(out, &models.TournamentTeam{
			ID:     id,
			Seed:   seedVal,
			Region: regionVal,
			School: &models.School{Name: schoolName},
		})
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterating teams: %w", rows.Err())
	}
	if len(out) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(out))
	}
	return out, nil
}

// loadGameOutcomeSpecFromPredictionBatch queries the latest prediction batch
// for the tournament and returns its game_outcome_spec_json as a winprob.Model.
func (s *Service) loadGameOutcomeSpecFromPredictionBatch(ctx context.Context, coreTournamentID string) (*winprob.Model, error) {
	var specJSON []byte
	err := s.pool.QueryRow(ctx, `
		SELECT game_outcome_spec_json
		FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid
			AND game_outcome_spec_json IS NOT NULL
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, coreTournamentID).Scan(&specJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no prediction batch with game_outcome_spec found for tournament_id=%s", coreTournamentID)
		}
		return nil, fmt.Errorf("querying prediction batch spec: %w", err)
	}

	var spec winprob.Model
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		return nil, fmt.Errorf("unmarshalling game_outcome_spec_json: %w", err)
	}
	return &spec, nil
}

func (s *Service) lockInFirstFourResults(
	ctx context.Context,
	br *models.BracketStructure,
	probs map[MatchupKey]float64,
) error {
	if br == nil {
		return errors.New("bracket must not be nil")
	}
	if probs == nil {
		return errors.New("probs must not be nil")
	}

	for _, g := range br.Games {
		if g == nil {
			continue
		}
		if g.Round != models.RoundFirstFour {
			continue
		}
		if g.Team1 == nil || g.Team2 == nil {
			continue
		}
		team1 := g.Team1.TeamID
		team2 := g.Team2.TeamID
		if team1 == "" || team2 == "" {
			continue
		}

		wins1, elim1, err := s.loadCoreTeamWinsEliminated(ctx, team1)
		if err != nil {
			return fmt.Errorf("loading team wins for first four team %s: %w", team1, err)
		}
		wins2, elim2, err := s.loadCoreTeamWinsEliminated(ctx, team2)
		if err != nil {
			return fmt.Errorf("loading team wins for first four team %s: %w", team2, err)
		}

		winner := ""
		if elim1 && !elim2 {
			winner = team2
		} else if elim2 && !elim1 {
			winner = team1
		} else if wins1 > wins2 {
			winner = team1
		} else if wins2 > wins1 {
			winner = team2
		} else {
			return fmt.Errorf("post_first_four requested but first four game not resolved for game_id=%s", g.GameID)
		}

		p1 := 0.0
		if winner == team1 {
			p1 = 1.0
			g.Winner = g.Team1
		} else {
			p1 = 0.0
			g.Winner = g.Team2
		}

		probs[MatchupKey{GameID: g.GameID, Team1ID: team1, Team2ID: team2}] = p1
		probs[MatchupKey{GameID: g.GameID, Team1ID: team2, Team2ID: team1}] = 1.0 - p1
	}

	return nil
}

func (s *Service) loadCoreTeamWinsEliminated(ctx context.Context, coreTeamID string) (int, bool, error) {
	var wins int
	var isEliminated bool
	err := s.pool.QueryRow(ctx, `
		SELECT wins, is_eliminated
		FROM core.teams
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, coreTeamID).Scan(&wins, &isEliminated)
	if err != nil {
		return 0, false, fmt.Errorf("querying team wins for %s: %w", coreTeamID, err)
	}
	return wins, isEliminated, nil
}
