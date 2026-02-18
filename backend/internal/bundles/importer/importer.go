package importer

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Report struct {
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	DryRun          bool      `json:"dry_run"`
	Schools         int       `json:"schools"`
	Tournaments     int       `json:"tournaments"`
	TournamentTeams int       `json:"tournament_teams"`
	Calcuttas       int       `json:"calcuttas"`
	Entries         int       `json:"entries"`
	Bids            int       `json:"bids"`
	Payouts         int       `json:"payouts"`
	Rounds          int       `json:"rounds"`
}

type Options struct {
	DryRun bool
}

func ImportFromDir(ctx context.Context, pool *pgxpool.Pool, inDir string, opts Options) (Report, error) {
	report := Report{StartedAt: time.Now().UTC(), DryRun: opts.DryRun}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return report, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	counts, err := importAll(ctx, tx, inDir)
	if err != nil {
		return report, err
	}
	report.Schools = counts.schools
	report.Tournaments = counts.tournaments
	report.TournamentTeams = counts.teams
	report.Calcuttas = counts.calcuttas
	report.Entries = counts.entries
	report.Bids = counts.bids
	report.Payouts = counts.payouts
	report.Rounds = counts.rounds

	if opts.DryRun {
		report.FinishedAt = time.Now().UTC()
		return report, nil
	}
	if err := tx.Commit(ctx); err != nil {
		return report, err
	}
	committed = true
	report.FinishedAt = time.Now().UTC()
	return report, nil
}

type importCounts struct {
	schools     int
	tournaments int
	teams       int
	calcuttas   int
	entries     int
	bids        int
	payouts     int
	rounds      int
}

func importAll(ctx context.Context, tx pgx.Tx, inDir string) (importCounts, error) {
	var c importCounts

	sc, err := importSchools(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.schools = sc

	tc, teams, err := importTournaments(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.tournaments = tc
	c.teams = teams

	cc, entries, bids, payouts, rounds, err := importCalcuttas(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.calcuttas = cc
	c.entries = entries
	c.bids = bids
	c.payouts = payouts
	c.rounds = rounds

	return c, nil
}
