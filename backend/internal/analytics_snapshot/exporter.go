package analytics_snapshot

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Manifest struct {
	SchemaVersion int       `json:"schema_version"`
	ExportedAt    time.Time `json:"exported_at"`

	TournamentID   string `json:"tournament_id"`
	TournamentKey  string `json:"tournament_key"`
	TournamentName string `json:"tournament_name"`

	PoolID   string `json:"pool_id"`
	PoolKey  string `json:"pool_key"`
	PoolName string `json:"pool_name"`

	RowCounts map[string]int `json:"row_counts"`
	Notes     string         `json:"notes"`
}

type ExportResult struct {
	Manifest Manifest
}

type ExportInputs struct {
	Bracket *models.BracketStructure
}

func ExportToDir(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time, tournamentID string, poolID string, inputs ExportInputs) (ExportResult, error) {
	var res ExportResult

	tournamentKey, tournamentName, rounds, err := loadTournament(ctx, pool, tournamentID)
	if err != nil {
		return res, err
	}

	poolKey, poolName, err := loadPool(ctx, pool, poolID, tournamentID)
	if err != nil {
		return res, err
	}

	schoolSlugByID, teamsRows, err := loadTeams(ctx, pool, tournamentID)
	if err != nil {
		return res, err
	}

	portfoliosRows, err := loadPortfolios(ctx, pool, poolID)
	if err != nil {
		return res, err
	}

	portfolioInvestmentsRows, err := loadPortfolioInvestments(ctx, pool, poolID, tournamentKey, schoolSlugByID)
	if err != nil {
		return res, err
	}

	roundScoringRows, err := loadRoundScoring(ctx, pool, poolID)
	if err != nil {
		return res, err
	}

	payoutRows, err := loadPayouts(ctx, pool, poolID)
	if err != nil {
		return res, err
	}

	gamesRows, err := buildGamesRows(tournamentKey, inputs.Bracket, schoolSlugByID)
	if err != nil {
		return res, err
	}

	rowCounts := map[string]int{}

	if err := writeCSV(filepath.Join(outDir, "teams.csv"), teamsRows.header, teamsRows.rows); err != nil {
		return res, err
	}
	rowCounts["teams.csv"] = len(teamsRows.rows)

	if err := writeCSV(filepath.Join(outDir, "games.csv"), gamesRows.header, gamesRows.rows); err != nil {
		return res, err
	}
	rowCounts["games.csv"] = len(gamesRows.rows)

	if err := writeCSV(filepath.Join(outDir, "portfolios.csv"), portfoliosRows.header, portfoliosRows.rows); err != nil {
		return res, err
	}
	rowCounts["portfolios.csv"] = len(portfoliosRows.rows)

	if err := writeCSV(filepath.Join(outDir, "portfolio_investments.csv"), portfolioInvestmentsRows.header, portfolioInvestmentsRows.rows); err != nil {
		return res, err
	}
	rowCounts["portfolio_investments.csv"] = len(portfolioInvestmentsRows.rows)

	if err := writeCSV(filepath.Join(outDir, "round_scoring.csv"), roundScoringRows.header, roundScoringRows.rows); err != nil {
		return res, err
	}
	rowCounts["round_scoring.csv"] = len(roundScoringRows.rows)

	if err := writeCSV(filepath.Join(outDir, "payouts.csv"), payoutRows.header, payoutRows.rows); err != nil {
		return res, err
	}
	rowCounts["payouts.csv"] = len(payoutRows.rows)

	m := Manifest{
		SchemaVersion:  1,
		ExportedAt:     generatedAt,
		TournamentID:   tournamentID,
		TournamentKey:  tournamentKey,
		TournamentName: tournamentName,
		PoolID:         poolID,
		PoolKey:        poolKey,
		PoolName:       poolName,
		RowCounts:      rowCounts,
		Notes:          fmt.Sprintf("games.csv is derived from bracket builder using team region/seed and team progress (wins+byes+is_eliminated). Tournament rounds=%d.", rounds),
	}

	if err := bundles.WriteJSON(filepath.Join(outDir, "manifest.json"), m); err != nil {
		return res, err
	}

	res.Manifest = m
	return res, nil
}

type csvTable struct {
	header []string
	rows   [][]string
}

func writeCSV(path string, header []string, rows [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func loadTournament(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (key string, name string, rounds int, err error) {
	err = pool.QueryRow(ctx, `
		SELECT t.import_key, comp.name || ' (' || seas.year || ')' AS name, t.rounds
		FROM core.tournaments t
		JOIN core.competitions comp ON comp.id = t.competition_id
		JOIN core.seasons seas ON seas.id = t.season_id
		WHERE t.id = $1 AND t.deleted_at IS NULL
	`, tournamentID).Scan(&key, &name, &rounds)
	return
}

func loadPool(ctx context.Context, dbPool *pgxpool.Pool, poolID string, tournamentID string) (key string, name string, err error) {
	var tID string
	err = dbPool.QueryRow(ctx, `
		SELECT tournament_id, name
		FROM core.pools
		WHERE id = $1 AND deleted_at IS NULL
	`, poolID).Scan(&tID, &name)
	if err != nil {
		return "", "", err
	}
	if tID != tournamentID {
		return "", "", fmt.Errorf("pool %s does not belong to tournament %s", poolID, tournamentID)
	}
	key = "pool-" + poolID
	return key, name, nil
}

func loadTeams(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (map[string]string, csvTable, error) {
	r, err := pool.Query(ctx, `
		SELECT
			s.slug,
			s.name,
			t.school_id,
			t.seed,
			t.region,
			t.byes,
			t.wins,
			t.is_eliminated,
			kps.net_rtg,
			kps.o_rtg,
			kps.d_rtg,
			kps.adj_t,
			(tr.import_key)
		FROM core.teams t
		JOIN core.schools s ON s.id = t.school_id
		JOIN core.tournaments tr ON tr.id = t.tournament_id
		LEFT JOIN core.team_kenpom_stats kps ON kps.team_id = t.id AND kps.deleted_at IS NULL
		WHERE t.tournament_id = $1 AND t.deleted_at IS NULL AND s.deleted_at IS NULL AND tr.deleted_at IS NULL
		ORDER BY t.region ASC, t.seed ASC, s.slug ASC
	`, tournamentID)
	if err != nil {
		return nil, csvTable{}, err
	}
	defer r.Close()

	header := []string{"tournament_key", "team_key", "school_slug", "school_name", "seed", "region", "byes", "wins", "is_eliminated", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"}
	rows := make([][]string, 0)
	slugBySchoolID := make(map[string]string)

	for r.Next() {
		var (
			slug, schoolName, schoolID, region, tournamentKey string
			seed, byes, wins                                  int
			isEliminated                                      bool
			net, o, d, adj                                    *float64
		)
		if err := r.Scan(&slug, &schoolName, &schoolID, &seed, &region, &byes, &wins, &isEliminated, &net, &o, &d, &adj, &tournamentKey); err != nil {
			return nil, csvTable{}, err
		}
		slugBySchoolID[schoolID] = slug
		teamKey := tournamentKey + ":" + slug
		rows = append(rows, []string{
			tournamentKey,
			teamKey,
			slug,
			schoolName,
			strconv.Itoa(seed),
			region,
			strconv.Itoa(byes),
			strconv.Itoa(wins),
			strconv.FormatBool(isEliminated),
			formatFloatPtr(net),
			formatFloatPtr(o),
			formatFloatPtr(d),
			formatFloatPtr(adj),
		})
	}
	if err := r.Err(); err != nil {
		return nil, csvTable{}, err
	}

	return slugBySchoolID, csvTable{header: header, rows: rows}, nil
}

func loadPortfolios(ctx context.Context, dbPool *pgxpool.Pool, poolID string) (csvTable, error) {
	r, err := dbPool.Query(ctx, `
		SELECT id, name
		FROM core.portfolios
		WHERE pool_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, poolID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	header := []string{"pool_key", "portfolio_key", "portfolio_name"}
	rows := make([][]string, 0)

	poolKey := "pool-" + poolID

	for r.Next() {
		var id, name string
		if err := r.Scan(&id, &name); err != nil {
			return csvTable{}, err
		}
		portfolioKey := poolKey + ":portfolio-" + id
		rows = append(rows, []string{poolKey, portfolioKey, name})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

func loadPortfolioInvestments(ctx context.Context, dbPool *pgxpool.Pool, poolID string, tournamentKey string, schoolSlugByID map[string]string) (csvTable, error) {
	r, err := dbPool.Query(ctx, `
		SELECT
			inv.id,
			p.id,
			t.school_id,
			inv.credits
		FROM core.investments inv
		JOIN core.portfolios p ON p.id = inv.portfolio_id
		JOIN core.teams t ON t.id = inv.team_id
		WHERE p.pool_id = $1 AND inv.deleted_at IS NULL AND p.deleted_at IS NULL AND t.deleted_at IS NULL
		ORDER BY p.created_at ASC
	`, poolID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	poolKey := "pool-" + poolID
	header := []string{"pool_key", "portfolio_key", "team_key", "credits"}
	rows := make([][]string, 0)

	for r.Next() {
		var _investmentID, portfolioID, schoolID string
		var credits int
		if err := r.Scan(&_investmentID, &portfolioID, &schoolID, &credits); err != nil {
			return csvTable{}, err
		}
		pKey := poolKey + ":portfolio-" + portfolioID
		slug := schoolSlugByID[schoolID]
		teamKey := ""
		if slug != "" {
			teamKey = tournamentKey + ":" + slug
		} else {
			teamKey = tournamentKey + ":school-" + schoolID
		}
		rows = append(rows, []string{poolKey, pKey, teamKey, strconv.Itoa(credits)})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

func loadRoundScoring(ctx context.Context, dbPool *pgxpool.Pool, poolID string) (csvTable, error) {
	r, err := dbPool.Query(ctx, `
		SELECT r.win_index, r.points_awarded
		FROM core.pool_scoring_rules r
		WHERE r.pool_id = $1 AND r.deleted_at IS NULL
		ORDER BY r.win_index ASC
	`, poolID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	header := []string{"pool_key", "round", "points"}
	rows := make([][]string, 0)
	poolKey := "pool-" + poolID

	for r.Next() {
		var round, points int
		if err := r.Scan(&round, &points); err != nil {
			return csvTable{}, err
		}
		rows = append(rows, []string{poolKey, strconv.Itoa(round), strconv.Itoa(points)})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

func loadPayouts(ctx context.Context, dbPool *pgxpool.Pool, poolID string) (csvTable, error) {
	r, err := dbPool.Query(ctx, `
		SELECT p.position, p.amount_cents
		FROM core.payouts p
		WHERE p.pool_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.position ASC
	`, poolID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	header := []string{"pool_key", "position", "amount_cents"}
	rows := make([][]string, 0)
	poolKey := "pool-" + poolID

	for r.Next() {
		var pos, amount int
		if err := r.Scan(&pos, &amount); err != nil {
			return csvTable{}, err
		}
		rows = append(rows, []string{poolKey, strconv.Itoa(pos), strconv.Itoa(amount)})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

func buildGamesRows(tournamentKey string, bracket *models.BracketStructure, schoolSlugByID map[string]string) (csvTable, error) {
	header := []string{"tournament_key", "game_id", "round", "region", "sort_order", "team1_key", "team2_key", "winner_team_key", "next_game_id", "next_game_slot"}
	rows := make([][]string, 0)
	if bracket == nil {
		return csvTable{header: header, rows: rows}, nil
	}

	for _, g := range bracket.Games {
		team1Key := ""
		team2Key := ""
		winnerKey := ""

		if g.Team1 != nil {
			slug := schoolSlugByID[g.Team1.SchoolID]
			if slug != "" {
				team1Key = tournamentKey + ":" + slug
			}
		}
		if g.Team2 != nil {
			slug := schoolSlugByID[g.Team2.SchoolID]
			if slug != "" {
				team2Key = tournamentKey + ":" + slug
			}
		}
		if g.Winner != nil {
			slug := schoolSlugByID[g.Winner.SchoolID]
			if slug != "" {
				winnerKey = tournamentKey + ":" + slug
			}
		}

		nextGameID := ""
		if g.NextGameID != "" {
			nextGameID = g.NextGameID
		}
		nextGameSlot := ""
		if g.NextGameSlot != 0 {
			nextGameSlot = strconv.Itoa(g.NextGameSlot)
		}

		rows = append(rows, []string{
			tournamentKey,
			g.GameID,
			string(g.Round),
			g.Region,
			strconv.Itoa(g.SortOrder),
			team1Key,
			team2Key,
			winnerKey,
			nextGameID,
			nextGameSlot,
		})
	}

	return csvTable{header: header, rows: rows}, nil
}

func formatFloatPtr(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}
