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

func importCalcuttas(ctx context.Context, tx pgx.Tx, inDir string) (int, int, int, int, int, error) {
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
			return 0, 0, 0, 0, 0, nil
		}
		return 0, 0, 0, 0, 0, err
	}
	sort.Strings(paths)

	entryCount := 0
	bidCount := 0
	payoutCount := 0
	roundCount := 0

	for _, path := range paths {
		var b bundles.CalcuttaBundle
		if err := readJSON(path, &b); err != nil {
			return 0, 0, 0, 0, 0, err
		}

		var tournamentID string
		err := tx.QueryRow(ctx, `
			SELECT id
			FROM core.tournaments
			WHERE import_key = $1 AND deleted_at IS NULL
		`, b.Tournament.ImportKey).Scan(&tournamentID)
		if err != nil {
			return 0, 0, 0, 0, 0, fmt.Errorf("tournament import_key %s not found: %w", b.Tournament.ImportKey, err)
		}

		ownerID, err := ensureUser(ctx, tx, b.Calcutta.Owner, b.Calcutta.Key)
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}

		calcuttaID := strings.TrimSpace(b.Calcutta.LegacyID)
		if calcuttaID == "" {
			err = tx.QueryRow(ctx, `
				INSERT INTO core.calcuttas (tournament_id, owner_id, name)
				VALUES ($1, $2, $3)
				RETURNING id
			`, tournamentID, ownerID, b.Calcutta.Name).Scan(&calcuttaID)
		} else {
			err = tx.QueryRow(ctx, `
				INSERT INTO core.calcuttas (id, tournament_id, owner_id, name)
				VALUES ($1::uuid, $2, $3, $4)
				ON CONFLICT (id)
				DO UPDATE SET
					tournament_id = EXCLUDED.tournament_id,
					owner_id = EXCLUDED.owner_id,
					name = EXCLUDED.name,
					updated_at = NOW(),
					deleted_at = NULL
				RETURNING id
			`, calcuttaID, tournamentID, ownerID, b.Calcutta.Name).Scan(&calcuttaID)
		}
		if err != nil {
			return 0, 0, 0, 0, 0, err
		}

		for _, r := range b.Rounds {
			// Canonical scoring rules
			_, err := tx.Exec(ctx, `
				INSERT INTO core.calcutta_scoring_rules (calcutta_id, win_index, points_awarded)
				VALUES ($1, $2, $3)
				ON CONFLICT (calcutta_id, win_index)
				DO UPDATE SET points_awarded = EXCLUDED.points_awarded, updated_at = NOW(), deleted_at = NULL
			`, calcuttaID, r.Round, r.Points)
			if err != nil {
				return 0, 0, 0, 0, 0, err
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
				return 0, 0, 0, 0, 0, err
			}
			payoutCount++
		}

		entryIDByKey := make(map[string]string, len(b.Entries))
		for _, e := range b.Entries {
			var entryUserID *string
			if e.UserEmail != nil {
				uid, err := ensureUserByEmail(ctx, tx, *e.UserEmail, e.UserName)
				if err != nil {
					return 0, 0, 0, 0, 0, err
				}
				entryUserID = &uid
			}

			var entryID string
			legacyEntryID := strings.TrimSpace(e.LegacyID)
			if legacyEntryID == "" {
				err := tx.QueryRow(ctx, `
					INSERT INTO core.entries (name, user_id, calcutta_id)
					VALUES ($1, $2, $3)
					RETURNING id
				`, e.Name, entryUserID, calcuttaID).Scan(&entryID)
				if err != nil {
					return 0, 0, 0, 0, 0, err
				}
			} else {
				err := tx.QueryRow(ctx, `
					INSERT INTO core.entries (id, name, user_id, calcutta_id)
					VALUES ($1::uuid, $2, $3, $4)
					ON CONFLICT (id)
					DO UPDATE SET name = EXCLUDED.name, user_id = EXCLUDED.user_id, calcutta_id = EXCLUDED.calcutta_id, updated_at = NOW(), deleted_at = NULL
					RETURNING id
				`, legacyEntryID, e.Name, entryUserID, calcuttaID).Scan(&entryID)
				if err != nil {
					return 0, 0, 0, 0, 0, err
				}
			}
			entryIDByKey[e.Key] = entryID
			entryCount++
		}

		for _, bid := range b.Bids {
			entryID := entryIDByKey[bid.EntryKey]
			if entryID == "" {
				return 0, 0, 0, 0, 0, fmt.Errorf("bid references unknown entry_key %s", bid.EntryKey)
			}

			var teamID string
			err := tx.QueryRow(ctx, `
				SELECT t.id
				FROM core.teams t
				JOIN core.schools s ON s.id = t.school_id
				WHERE t.tournament_id = $1 AND s.slug = $2 AND t.deleted_at IS NULL AND s.deleted_at IS NULL
			`, tournamentID, bid.SchoolSlug).Scan(&teamID)
			if err != nil {
				return 0, 0, 0, 0, 0, fmt.Errorf("tournament team not found for tournament %s school %s: %w", b.Tournament.ImportKey, bid.SchoolSlug, err)
			}

			legacyEntryTeamID := strings.TrimSpace(bid.LegacyEntryTeamID)
			if legacyEntryTeamID == "" {
				_, err = tx.Exec(ctx, `
					INSERT INTO core.entry_teams (entry_id, team_id, bid_points)
					VALUES ($1, $2, $3)
				`, entryID, teamID, bid.Bid)
			} else {
				_, err = tx.Exec(ctx, `
					INSERT INTO core.entry_teams (id, entry_id, team_id, bid_points)
					VALUES ($1::uuid, $2, $3, $4)
					ON CONFLICT (id)
					DO UPDATE SET entry_id = EXCLUDED.entry_id, team_id = EXCLUDED.team_id, bid_points = EXCLUDED.bid_points, updated_at = NOW(), deleted_at = NULL
				`, legacyEntryTeamID, entryID, teamID, bid.Bid)
			}
			if err != nil {
				return 0, 0, 0, 0, 0, err
			}
			bidCount++
		}
	}

	return len(paths), entryCount, bidCount, payoutCount, roundCount, nil
}

func ensureUser(ctx context.Context, tx pgx.Tx, u *bundles.UserRef, fallbackKey string) (string, error) {
	email := ""
	first := ""
	last := ""
	if u != nil {
		if u.Email != nil {
			email = strings.TrimSpace(*u.Email)
		}
		if u.FirstName != nil {
			first = strings.TrimSpace(*u.FirstName)
		}
		if u.LastName != nil {
			last = strings.TrimSpace(*u.LastName)
		}
	}
	if email == "" {
		email = fmt.Sprintf("owner+%s@local", strings.ReplaceAll(fallbackKey, " ", "-"))
	}
	if first == "" {
		first = "Unknown"
	}
	if last == "" {
		last = ""
	}
	full := strings.TrimSpace(first + " " + last)
	var fullName *string
	if full != "" {
		fullName = &full
	}
	return ensureUserByEmail(ctx, tx, email, fullName)
}

func ensureUserByEmail(ctx context.Context, tx pgx.Tx, email string, fullName *string) (string, error) {
	first := ""
	last := ""
	if fullName != nil {
		parts := strings.Fields(*fullName)
		if len(parts) > 0 {
			first = parts[0]
		}
		if len(parts) > 1 {
			last = strings.Join(parts[1:], " ")
		}
	}
	if first == "" {
		first = "Unknown"
	}
	if last == "" {
		last = ""
	}

	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO core.users (email, first_name, last_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (email)
		DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name, updated_at = NOW()
		RETURNING id
	`, email, first, last).Scan(&id)
	return id, err
}
