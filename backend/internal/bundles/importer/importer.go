package importer

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Report struct {
	StartedAt time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	DryRun    bool      `json:"dryRun"`
	Calcuttas int       `json:"calcuttas"`
	Entries   int       `json:"entries"`
	StubUsers int       `json:"stubUsers"`
	Bids      int       `json:"bids"`
	Payouts   int       `json:"payouts"`
	Rounds    int       `json:"rounds"`
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
	report.Calcuttas = counts.calcuttas
	report.Entries = counts.entries
	report.StubUsers = counts.stubUsers
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
	calcuttas int
	entries   int
	stubUsers int
	bids      int
	payouts   int
	rounds    int
}

func importAll(ctx context.Context, tx pgx.Tx, inDir string) (importCounts, error) {
	var c importCounts

	cc, entries, stubUsers, bids, payouts, rounds, err := importCalcuttas(ctx, tx, inDir)
	if err != nil {
		return c, err
	}
	c.calcuttas = cc
	c.entries = entries
	c.stubUsers = stubUsers
	c.bids = bids
	c.payouts = payouts
	c.rounds = rounds

	return c, nil
}
