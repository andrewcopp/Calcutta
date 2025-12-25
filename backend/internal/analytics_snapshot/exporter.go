package analytics_snapshot

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Manifest struct {
	SchemaVersion int       `json:"schema_version"`
	ExportedAt    time.Time `json:"exported_at"`

	TournamentID   string `json:"tournament_id"`
	TournamentKey  string `json:"tournament_key"`
	TournamentName string `json:"tournament_name"`

	CalcuttaID   string `json:"calcutta_id"`
	CalcuttaKey  string `json:"calcutta_key"`
	CalcuttaName string `json:"calcutta_name"`

	RowCounts map[string]int `json:"row_counts"`
	Notes     string         `json:"notes"`
}

type ExportResult struct {
	Manifest Manifest
}

type ExportInputs struct {
	Bracket *models.BracketStructure
}

func ExportToDir(ctx context.Context, pool *pgxpool.Pool, outDir string, generatedAt time.Time, tournamentID string, calcuttaID string, inputs ExportInputs) (ExportResult, error) {
	var res ExportResult

	tournamentKey, tournamentName, rounds, err := loadTournament(ctx, pool, tournamentID)
	if err != nil {
		return res, err
	}

	calcuttaKey, calcuttaName, err := loadCalcutta(ctx, pool, calcuttaID, tournamentID)
	if err != nil {
		return res, err
	}

	schoolSlugByID, teamsRows, err := loadTeams(ctx, pool, tournamentID)
	if err != nil {
		return res, err
	}

	entriesRows, err := loadEntries(ctx, pool, calcuttaID)
	if err != nil {
		return res, err
	}

	entryBidsRows, err := loadEntryBids(ctx, pool, calcuttaID, tournamentID, tournamentKey, schoolSlugByID)
	if err != nil {
		return res, err
	}

	roundScoringRows, err := loadRoundScoring(ctx, pool, calcuttaID)
	if err != nil {
		return res, err
	}

	payoutRows, err := loadPayouts(ctx, pool, calcuttaID)
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

	if err := writeCSV(filepath.Join(outDir, "entries.csv"), entriesRows.header, entriesRows.rows); err != nil {
		return res, err
	}
	rowCounts["entries.csv"] = len(entriesRows.rows)

	if err := writeCSV(filepath.Join(outDir, "entry_bids.csv"), entryBidsRows.header, entryBidsRows.rows); err != nil {
		return res, err
	}
	rowCounts["entry_bids.csv"] = len(entryBidsRows.rows)

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
		CalcuttaID:     calcuttaID,
		CalcuttaKey:    calcuttaKey,
		CalcuttaName:   calcuttaName,
		RowCounts:      rowCounts,
		Notes:          fmt.Sprintf("games.csv is derived from bracket builder using team region/seed and team progress (wins+byes+eliminated). Tournament rounds=%d.", rounds),
	}

	if err := writeJSON(filepath.Join(outDir, "manifest.json"), m); err != nil {
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

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func loadTournament(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (key string, name string, rounds int, err error) {
	err = pool.QueryRow(ctx, `
		SELECT import_key, name, rounds
		FROM tournaments
		WHERE id = $1 AND deleted_at IS NULL
	`, tournamentID).Scan(&key, &name, &rounds)
	return
}

func loadCalcutta(ctx context.Context, pool *pgxpool.Pool, calcuttaID string, tournamentID string) (key string, name string, err error) {
	var tID string
	err = pool.QueryRow(ctx, `
		SELECT tournament_id, COALESCE(key, ''), name
		FROM calcuttas
		WHERE id = $1 AND deleted_at IS NULL
	`, calcuttaID).Scan(&tID, &key, &name)
	if err != nil {
		return "", "", err
	}
	if tID != tournamentID {
		return "", "", fmt.Errorf("calcutta %s does not belong to tournament %s", calcuttaID, tournamentID)
	}
	if key == "" {
		key = "calcutta-" + calcuttaID
	}
	return key, name, nil
}

type teamsLoadResult struct {
	schoolSlugByID map[string]string
	table          csvTable
}

func loadTeams(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (map[string]string, csvTable, error) {
	r, err := pool.Query(ctx, `
		SELECT
			s.slug,
			s.name,
			tt.school_id,
			tt.seed,
			tt.region,
			tt.byes,
			tt.wins,
			tt.eliminated,
			kps.net_rtg,
			kps.o_rtg,
			kps.d_rtg,
			kps.adj_t,
			(t.import_key)
		FROM tournament_teams tt
		JOIN schools s ON s.id = tt.school_id
		JOIN tournaments t ON t.id = tt.tournament_id
		LEFT JOIN tournament_team_kenpom_stats kps ON kps.tournament_team_id = tt.id AND kps.deleted_at IS NULL
		WHERE tt.tournament_id = $1 AND tt.deleted_at IS NULL AND s.deleted_at IS NULL AND t.deleted_at IS NULL
		ORDER BY tt.region ASC, tt.seed ASC, s.slug ASC
	`, tournamentID)
	if err != nil {
		return nil, csvTable{}, err
	}
	defer r.Close()

	header := []string{"tournament_key", "team_key", "school_slug", "school_name", "seed", "region", "byes", "wins", "eliminated", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"}
	rows := make([][]string, 0)
	slugBySchoolID := make(map[string]string)

	for r.Next() {
		var (
			slug, schoolName, schoolID, region, tournamentKey string
			seed, byes, wins                                  int
			eliminated                                        bool
			net, o, d, adj                                    *float64
		)
		if err := r.Scan(&slug, &schoolName, &schoolID, &seed, &region, &byes, &wins, &eliminated, &net, &o, &d, &adj, &tournamentKey); err != nil {
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
			strconv.FormatBool(eliminated),
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

func loadEntries(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (csvTable, error) {
	r, err := pool.Query(ctx, `
		SELECT COALESCE(key, ''), id, name
		FROM calcutta_entries
		WHERE calcutta_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, calcuttaID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	header := []string{"calcutta_key", "entry_key", "entry_name"}
	rows := make([][]string, 0)

	var calcuttaKey string
	err = pool.QueryRow(ctx, `SELECT COALESCE(key, '') FROM calcuttas WHERE id = $1 AND deleted_at IS NULL`, calcuttaID).Scan(&calcuttaKey)
	if err != nil {
		return csvTable{}, err
	}
	if calcuttaKey == "" {
		calcuttaKey = "calcutta-" + calcuttaID
	}

	for r.Next() {
		var key, id, name string
		if err := r.Scan(&key, &id, &name); err != nil {
			return csvTable{}, err
		}
		if key == "" {
			key = calcuttaKey + ":entry-" + id
		}
		rows = append(rows, []string{calcuttaKey, key, name})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

func loadEntryBids(ctx context.Context, pool *pgxpool.Pool, calcuttaID string, tournamentID string, tournamentKey string, schoolSlugByID map[string]string) (csvTable, error) {
	r, err := pool.Query(ctx, `
		SELECT
			COALESCE(c.key, ''),
			COALESCE(e.key, ''),
			e.id,
			tt.school_id,
			cet.bid
		FROM calcutta_entry_teams cet
		JOIN calcutta_entries e ON e.id = cet.entry_id
		JOIN calcuttas c ON c.id = e.calcutta_id
		JOIN tournament_teams tt ON tt.id = cet.team_id
		WHERE c.id = $1 AND c.deleted_at IS NULL AND e.deleted_at IS NULL AND cet.deleted_at IS NULL AND tt.deleted_at IS NULL
		ORDER BY e.created_at ASC
	`, calcuttaID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	calcuttaKey := ""
	header := []string{"calcutta_key", "entry_key", "team_key", "bid_amount"}
	rows := make([][]string, 0)

	for r.Next() {
		var cKey, eKey, entryID, schoolID string
		var bid int
		if err := r.Scan(&cKey, &eKey, &entryID, &schoolID, &bid); err != nil {
			return csvTable{}, err
		}
		if calcuttaKey == "" {
			calcuttaKey = cKey
			if calcuttaKey == "" {
				calcuttaKey = "calcutta-" + calcuttaID
			}
		}
		if eKey == "" {
			eKey = calcuttaKey + ":entry-" + entryID
		}
		slug := schoolSlugByID[schoolID]
		teamKey := ""
		if slug != "" {
			teamKey = tournamentKey + ":" + slug
		} else {
			teamKey = tournamentKey + ":school-" + schoolID
		}
		rows = append(rows, []string{calcuttaKey, eKey, teamKey, strconv.Itoa(bid)})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	_ = tournamentID
	return csvTable{header: header, rows: rows}, nil
}

func loadRoundScoring(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (csvTable, error) {
	r, err := pool.Query(ctx, `
		SELECT COALESCE(c.key, ''), r.round, r.points
		FROM calcutta_rounds r
		JOIN calcuttas c ON c.id = r.calcutta_id
		WHERE r.calcutta_id = $1 AND r.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY r.round ASC
	`, calcuttaID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	header := []string{"calcutta_key", "round", "points"}
	rows := make([][]string, 0)

	for r.Next() {
		var calcuttaKey string
		var round, points int
		if err := r.Scan(&calcuttaKey, &round, &points); err != nil {
			return csvTable{}, err
		}
		if calcuttaKey == "" {
			calcuttaKey = "calcutta-" + calcuttaID
		}
		rows = append(rows, []string{calcuttaKey, strconv.Itoa(round), strconv.Itoa(points)})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

func loadPayouts(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (csvTable, error) {
	r, err := pool.Query(ctx, `
		SELECT COALESCE(c.key, ''), p.position, p.amount_cents
		FROM calcutta_payouts p
		JOIN calcuttas c ON c.id = p.calcutta_id
		WHERE p.calcutta_id = $1 AND p.deleted_at IS NULL AND c.deleted_at IS NULL
		ORDER BY p.position ASC
	`, calcuttaID)
	if err != nil {
		return csvTable{}, err
	}
	defer r.Close()

	header := []string{"calcutta_key", "position", "amount_cents"}
	rows := make([][]string, 0)

	for r.Next() {
		var calcuttaKey string
		var pos, amount int
		if err := r.Scan(&calcuttaKey, &pos, &amount); err != nil {
			return csvTable{}, err
		}
		if calcuttaKey == "" {
			calcuttaKey = "calcutta-" + calcuttaID
		}
		rows = append(rows, []string{calcuttaKey, strconv.Itoa(pos), strconv.Itoa(amount)})
	}
	if err := r.Err(); err != nil {
		return csvTable{}, err
	}

	return csvTable{header: header, rows: rows}, nil
}

type gamesTable struct {
	header []string
	rows   [][]string
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
