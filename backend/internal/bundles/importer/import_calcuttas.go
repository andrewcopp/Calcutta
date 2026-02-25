package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/jackc/pgx/v5"
)

func importCalcuttas(ctx context.Context, tx pgx.Tx, inDir string) (int, int, int, int, int, int, error) {
	root := filepath.Join(inDir, "calcuttas")

	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, 0, 0, 0, 0, nil
		}
		return 0, 0, 0, 0, 0, 0, err
	}
	sort.Strings(paths)

	entryCount := 0
	stubUserCount := 0
	bidCount := 0
	payoutCount := 0
	roundCount := 0

	for _, path := range paths {
		var b bundles.CalcuttaBundle
		if err := bundles.ReadJSON(path, &b); err != nil {
			return 0, 0, 0, 0, 0, 0, err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			SELECT id
			FROM core.tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, fmt.Errorf("tournament import_key %q not found — have you run seed migrations? %w", b.Tournament.ImportKey, err)
		}

		ownerName := ownerDisplayName(b.Calcutta.Owner, b.Calcutta.Key)
		ownerID, err := createStubUser(ctx, tx, ownerName)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, err
		}
		stubUserCount++

		var calcuttaID string
		err = tx.QueryRow(ctx, `
			INSERT INTO core.calcuttas (tournament_id, owner_id, created_by, name)
			VALUES ($1, $2, $2, $3)
			RETURNING id
		`, tournamentID, ownerID, b.Calcutta.Name).Scan(&calcuttaID)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, err
		}

		for _, r := range b.Rounds {
			_, err := tx.Exec(ctx, `
				INSERT INTO core.calcutta_scoring_rules (calcutta_id, win_index, points_awarded)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, win_index)
				DO UPDATE SET points_awarded = EXCLUDED.points_awarded, updated_at = NOW(), deleted_at = NULL
			`, calcuttaID, r.Round, r.Points)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			roundCount++
		}

		for _, p := range b.Payouts {
			_, err := tx.Exec(ctx, `
				INSERT INTO core.payouts (calcutta_id, position, amount_cents)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, position)
				DO UPDATE SET amount_cents = EXCLUDED.amount_cents, updated_at = NOW(), deleted_at = NULL
			`, calcuttaID, p.Position, p.AmountCents)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			payoutCount++
		}

		entryIDByKey := make(map[string]string, len(b.Entries))
		for _, e := range b.Entries {
			// Determine display name: prefer UserName, fallback to entry Name
			displayName := e.Name
			if e.UserName != nil && strings.TrimSpace(*e.UserName) != "" {
				displayName = *e.UserName
			}

			entryUserID, err := createStubUser(ctx, tx, displayName)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			stubUserCount++

			var entryID string
			err = tx.QueryRow(ctx, `
				INSERT INTO core.entries (name, user_id, calcutta_id)
				VALUES ($1, $2, $3)
				RETURNING id
			`, e.Name, entryUserID, calcuttaID).Scan(&entryID)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			entryIDByKey[e.Key] = entryID
			entryCount++
		}

		for _, bid := range b.Bids {
			entryID := entryIDByKey[bid.EntryKey]
			if entryID == "" {
				return 0, 0, 0, 0, 0, 0, fmt.Errorf("bid references unknown entry_key %s", bid.EntryKey)
			}

			var teamID string
			err := tx.QueryRow(ctx, `
				SELECT t.id
				FROM core.teams t
				JOIN core.schools s ON s.id = t.school_id
				WHERE t.tournament_id = $1 AND s.slug = $2 AND t.deleted_at IS NULL AND s.deleted_at IS NULL
			`, tournamentID, bid.SchoolSlug).Scan(&teamID)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, fmt.Errorf("team not found for tournament %q school %q — have you run seed migrations? %w", b.Tournament.ImportKey, bid.SchoolSlug, err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO core.entry_teams (entry_id, team_id, bid_points)
				VALUES ($1, $2, $3)
			`, entryID, teamID, bid.Bid)
			if err != nil {
				return 0, 0, 0, 0, 0, 0, err
			}
			bidCount++
		}
	}

	return len(paths), entryCount, stubUserCount, bidCount, payoutCount, roundCount, nil
}

// createStubUser always creates a new stub user (never deduplicates).
func createStubUser(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Unknown"
	}

	parts := strings.Fields(name)
	first := parts[0]
	last := ""
	if len(parts) > 1 {
		last = strings.Join(parts[1:], " ")
	}

	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO core.users (first_name, last_name, status, external_provider)
		VALUES ($1, $2, 'stub', 'historical')
		RETURNING id
	`, first, last).Scan(&id)
	return id, err
}

// ownerDisplayName extracts a display name from a UserRef, falling back to the calcutta key.
func ownerDisplayName(u *bundles.UserRef, fallbackKey string) string {
	if u == nil {
		return fallbackKey
	}
	first := ""
	last := ""
	if u.FirstName != nil {
		first = strings.TrimSpace(*u.FirstName)
	}
	if u.LastName != nil {
		last = strings.TrimSpace(*u.LastName)
	}
	fullName := strings.TrimSpace(first + " " + last)
	if fullName == "" {
		return fallbackKey
	}
	return fullName
}
