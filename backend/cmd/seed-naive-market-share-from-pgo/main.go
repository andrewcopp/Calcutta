package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sort"
	"time"

	tsim "github.com/andrewcopp/Calcutta/backend/internal/app/tournament_simulation"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/features/bracket"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type calcuttaContext struct {
	CalcuttaID       string
	CoreTournamentID string
}

type scoringRule struct {
	WinIndex      int
	PointsAwarded int
}

type roundReach struct {
	ReachFirstFour float64
	ReachR64       float64
	ReachR32       float64
	ReachS16       float64
	ReachE8        float64
	ReachFF        float64
	ReachChamp     float64
	WinChamp       float64
}

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var calcuttaID string
	var dryRun bool

	flag.StringVar(&calcuttaID, "calcutta-id", "", "Calcutta ID (uuid)")
	flag.BoolVar(&dryRun, "dry-run", false, "If set, compute and print but do not write predicted_market_share")
	flag.Parse()

	if calcuttaID == "" {
		flag.Usage()
		return fmt.Errorf("--calcutta-id is required")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database (pgxpool): %w", err)
	}
	defer pool.Close()

	ctx := context.Background()

	cc, err := loadCalcuttaContext(ctx, pool, calcuttaID)
	if err != nil {
		return err
	}

	scoringRules, err := loadScoringRules(ctx, pool, cc.CalcuttaID)
	if err != nil {
		return err
	}
	if len(scoringRules) == 0 {
		return fmt.Errorf("no calcutta scoring rules found")
	}

	ff, err := loadFinalFourConfig(ctx, pool, cc.CoreTournamentID)
	if err != nil {
		return err
	}

	teams, err := loadTeams(ctx, pool, cc.CoreTournamentID)
	if err != nil {
		return err
	}

	builder := appbracket.NewBracketBuilder()
	br, err := builder.BuildBracket(cc.CoreTournamentID, teams, ff)
	if err != nil {
		return err
	}

	probs, nPred, err := loadPredictedGameOutcomes(ctx, pool, cc.CoreTournamentID)
	if err != nil {
		return err
	}
	if nPred == 0 {
		return fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", cc.CoreTournamentID)
	}

	evByTeam, reachByTeam, err := computeExpectedValueFromPGO(br, probs, scoringRules)
	if err != nil {
		return err
	}

	// Print top 10 by EV for sanity
	type row struct {
		TeamID string
		EV     float64
		R      roundReach
	}
	rows := make([]row, 0, len(evByTeam))
	for tid, ev := range evByTeam {
		rows = append(rows, row{TeamID: tid, EV: ev, R: reachByTeam[tid]})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].EV > rows[j].EV })

	log.Printf("Computed EV from predicted_game_outcomes: teams=%d (showing top 10)", len(rows))
	for i := 0; i < len(rows) && i < 10; i++ {
		r := rows[i]
		log.Printf("%02d team_id=%s ev=%.4f reach_r64=%.3f reach_r32=%.3f reach_s16=%.3f reach_e8=%.3f reach_ff=%.3f reach_champ=%.3f win_champ=%.3f", i+1, r.TeamID, r.EV, r.R.ReachR64, r.R.ReachR32, r.R.ReachS16, r.R.ReachE8, r.R.ReachFF, r.R.ReachChamp, r.R.WinChamp)
	}

	if dryRun {
		log.Printf("dry-run: not writing derived.predicted_market_share")
		return nil
	}

	inserted, err := writeNaivePredictedMarketShare(ctx, pool, cc.CoreTournamentID, evByTeam)
	if err != nil {
		return err
	}
	log.Printf("seeded derived.predicted_market_share (tournament-scoped, run_id NULL): inserted=%d tournament_id=%s", inserted, cc.CoreTournamentID)
	return nil
}

func loadCalcuttaContext(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (*calcuttaContext, error) {
	var out calcuttaContext
	if err := pool.QueryRow(ctx, `
		SELECT c.id, t.id
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`, calcuttaID).Scan(&out.CalcuttaID, &out.CoreTournamentID); err != nil {
		return nil, err
	}
	return &out, nil
}

func loadScoringRules(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) ([]scoringRule, error) {
	rows, err := pool.Query(ctx, `
		SELECT win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY win_index ASC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]scoringRule, 0)
	for rows.Next() {
		var r scoringRule
		if err := rows.Scan(&r.WinIndex, &r.PointsAwarded); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func loadFinalFourConfig(ctx context.Context, pool *pgxpool.Pool, coreTournamentID string) (*models.FinalFourConfig, error) {
	var tl, bl, tr, br *string
	if err := pool.QueryRow(ctx, `
		SELECT final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
		FROM core.tournaments
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, coreTournamentID).Scan(&tl, &bl, &tr, &br); err != nil {
		return nil, err
	}

	cfg := &models.FinalFourConfig{}
	if tl != nil {
		cfg.TopLeftRegion = *tl
	}
	if bl != nil {
		cfg.BottomLeftRegion = *bl
	}
	if tr != nil {
		cfg.TopRightRegion = *tr
	}
	if br != nil {
		cfg.BottomRightRegion = *br
	}

	if cfg.TopLeftRegion == "" {
		cfg.TopLeftRegion = "East"
	}
	if cfg.BottomLeftRegion == "" {
		cfg.BottomLeftRegion = "West"
	}
	if cfg.TopRightRegion == "" {
		cfg.TopRightRegion = "South"
	}
	if cfg.BottomRightRegion == "" {
		cfg.BottomRightRegion = "Midwest"
	}

	return cfg, nil
}

func loadTeams(ctx context.Context, pool *pgxpool.Pool, coreTournamentID string) ([]*models.TournamentTeam, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			t.id,
			t.seed,
			t.region,
			s.name,
			s.id
		FROM core.teams t
		JOIN core.schools s
			ON s.id = t.school_id
			AND s.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
		ORDER BY t.seed ASC, s.name ASC
	`, coreTournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*models.TournamentTeam, 0)
	for rows.Next() {
		var id string
		var seed int
		var region string
		var schoolName string
		var schoolID string
		if err := rows.Scan(&id, &seed, &region, &schoolName, &schoolID); err != nil {
			return nil, err
		}
		out = append(out, &models.TournamentTeam{
			ID:     id,
			Seed:   seed,
			Region: region,
			School: &models.School{ID: schoolID, Name: schoolName},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(out) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(out))
	}
	return out, nil
}

func loadPredictedGameOutcomes(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (map[tsim.MatchupKey]float64, int, error) {
	rows, err := pool.Query(ctx, `
		SELECT game_id, team1_id, team2_id, p_team1_wins
		FROM derived.predicted_game_outcomes
		WHERE tournament_id = $1::uuid
			AND deleted_at IS NULL
	`, tournamentID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make(map[tsim.MatchupKey]float64)
	n := 0
	for rows.Next() {
		var gameID, t1, t2 string
		var p float64
		if err := rows.Scan(&gameID, &t1, &t2, &p); err != nil {
			return nil, 0, err
		}
		n++
		out[tsim.MatchupKey{GameID: gameID, Team1ID: t1, Team2ID: t2}] = p
		out[tsim.MatchupKey{GameID: gameID, Team1ID: t2, Team2ID: t1}] = 1.0 - p
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return out, n, nil
}

func computeExpectedValueFromPGO(
	br *models.BracketStructure,
	probs map[tsim.MatchupKey]float64,
	scoringRules []scoringRule,
) (map[string]float64, map[string]roundReach, error) {
	if br == nil {
		return nil, nil, errors.New("bracket must not be nil")
	}
	if len(br.Games) == 0 {
		return nil, nil, errors.New("bracket must have games")
	}
	if len(scoringRules) == 0 {
		return nil, nil, errors.New("scoringRules must not be empty")
	}

	games := make([]*models.BracketGame, 0, len(br.Games))
	prevByNext := make(map[string]map[int]string)
	for _, g := range br.Games {
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

	// gameID -> distribution of winners
	winnerDistByGame := make(map[string]map[string]float64)

	// teamID -> reach probabilities
	reach := make(map[string]roundReach)

	champGameID := ""

	for _, g := range games {
		if g == nil || g.GameID == "" {
			continue
		}
		if g.Round == models.RoundChampionship {
			champGameID = g.GameID
		}

		slot1 := make(map[string]float64)
		slot2 := make(map[string]float64)

		if g.Team1 != nil && g.Team1.TeamID != "" {
			slot1[g.Team1.TeamID] = 1.0
		} else {
			prev := prevByNext[g.GameID][1]
			if prev == "" {
				return nil, nil, fmt.Errorf("game %s missing Team1 and missing prev slot 1", g.GameID)
			}
			wd := winnerDistByGame[prev]
			if wd == nil {
				return nil, nil, fmt.Errorf("game %s slot 1 depends on %s but winner distribution not computed", g.GameID, prev)
			}
			for tid, p := range wd {
				slot1[tid] = p
			}
		}

		if g.Team2 != nil && g.Team2.TeamID != "" {
			slot2[g.Team2.TeamID] = 1.0
		} else {
			prev := prevByNext[g.GameID][2]
			if prev == "" {
				return nil, nil, fmt.Errorf("game %s missing Team2 and missing prev slot 2", g.GameID)
			}
			wd := winnerDistByGame[prev]
			if wd == nil {
				return nil, nil, fmt.Errorf("game %s slot 2 depends on %s but winner distribution not computed", g.GameID, prev)
			}
			for tid, p := range wd {
				slot2[tid] = p
			}
		}

		// Record reach probabilities for this round
		for tid, p := range slot1 {
			r := reach[tid]
			applyReach(&r, g.Round, p)
			reach[tid] = r
		}
		for tid, p := range slot2 {
			r := reach[tid]
			applyReach(&r, g.Round, p)
			reach[tid] = r
		}

		// Compute winner distribution
		winners := make(map[string]float64)
		for t1, pSlot1 := range slot1 {
			for t2, pSlot2 := range slot2 {
				pMatch := pSlot1 * pSlot2
				if pMatch == 0 {
					continue
				}
				p1, ok := probs[tsim.MatchupKey{GameID: g.GameID, Team1ID: t1, Team2ID: t2}]
				if !ok {
					return nil, nil, fmt.Errorf("missing predicted_game_outcomes for game_id=%s team1_id=%s team2_id=%s", g.GameID, t1, t2)
				}
				winners[t1] += pMatch * p1
				winners[t2] += pMatch * (1.0 - p1)
			}
		}

		winnerDistByGame[g.GameID] = winners
	}

	if champGameID == "" {
		return nil, nil, errors.New("championship game not found in bracket")
	}

	// winnerDist of championship game = P(win championship)
	champWinners := winnerDistByGame[champGameID]
	for tid, p := range champWinners {
		r := reach[tid]
		r.WinChamp = p
		reach[tid] = r
	}

	// Compute EV from scoring rules
	ev := make(map[string]float64)
	for tid, rr := range reach {
		// Scoring uses win_index <= (wins + byes).
		// Here we map P(progress >= k) to reach probabilities derived from bracket DP.
		// - progress>=1: in Round of 64 (non-FirstFour teams have byes=1 => always; FirstFour must win play-in)
		// - progress>=2: in Round of 32
		// - ...
		// - progress>=6: in Championship game
		// - progress>=7: wins Championship
		pAtLeast := func(winIndex int) float64 {
			switch winIndex {
			case 0:
				return 1.0
			case 1:
				return rr.ReachR64
			case 2:
				return rr.ReachR32
			case 3:
				return rr.ReachS16
			case 4:
				return rr.ReachE8
			case 5:
				return rr.ReachFF
			case 6:
				return rr.ReachChamp
			case 7:
				return rr.WinChamp
			default:
				// Not supported in this DP (would require modeling additional byes/rounds)
				return 0.0
			}
		}

		val := 0.0
		for _, sr := range scoringRules {
			val += float64(sr.PointsAwarded) * pAtLeast(sr.WinIndex)
		}
		ev[tid] = val
	}

	return ev, reach, nil
}

func applyReach(r *roundReach, round models.BracketRound, p float64) {
	switch round {
	case models.RoundFirstFour:
		r.ReachFirstFour += p
	case models.RoundOf64:
		r.ReachR64 += p
	case models.RoundOf32:
		r.ReachR32 += p
	case models.RoundSweet16:
		r.ReachS16 += p
	case models.RoundElite8:
		r.ReachE8 += p
	case models.RoundFinalFour:
		r.ReachFF += p
	case models.RoundChampionship:
		r.ReachChamp += p
	}
}

func writeNaivePredictedMarketShare(ctx context.Context, pool *pgxpool.Pool, tournamentID string, evByTeam map[string]float64) (int, error) {
	if tournamentID == "" {
		return 0, errors.New("tournamentID is required")
	}
	if len(evByTeam) == 0 {
		return 0, errors.New("evByTeam is empty")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		DELETE FROM derived.predicted_market_share
		WHERE tournament_id = $1::uuid
			AND calcutta_id IS NULL
			AND run_id IS NULL
	`, tournamentID)
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, v := range evByTeam {
		if v > 0 {
			total += v
		}
	}
	if total <= 0 {
		return 0, errors.New("total expected value is non-positive")
	}

	now := time.Now().UTC()
	inserted := 0
	for teamID, ev := range evByTeam {
		share := 0.0
		if ev > 0 {
			share = ev / total
		}
		predictedPoints := share * 100.0
		_, err := tx.Exec(ctx, `
			INSERT INTO derived.predicted_market_share (
				calcutta_id,
				tournament_id,
				team_id,
				predicted_share,
				predicted_points,
				created_at,
				updated_at
			)
			VALUES (NULL, $1::uuid, $2::uuid, $3, $4, $5, $5)
		`, tournamentID, teamID, share, predictedPoints, now)
		if err != nil {
			return 0, err
		}
		inserted++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return inserted, nil
}
