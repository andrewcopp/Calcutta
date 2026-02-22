package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultTournamentImportWorkerPollInterval = 2 * time.Second
	defaultTournamentImportWorkerStaleAfter   = 30 * time.Minute
)

type TournamentImportWorker struct {
	pool *pgxpool.Pool
}

func NewTournamentImportWorker(pool *pgxpool.Pool) *TournamentImportWorker {
	return &TournamentImportWorker{pool: pool}
}

func (w *TournamentImportWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultTournamentImportWorkerPollInterval, defaultTournamentImportWorkerStaleAfter)
}

func (w *TournamentImportWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		slog.Warn("tournament_import_worker_disabled", "reason", "database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultTournamentImportWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultTournamentImportWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			uploadID, ok, err := w.claimNextTournamentImport(ctx, staleAfter)
			if err != nil {
				slog.Error("tournament_import_claim_failed", "error", err)
				continue
			}
			if !ok {
				continue
			}
			w.processTournamentImport(ctx, uploadID)
		}
	}
}

func (w *TournamentImportWorker) claimNextTournamentImport(ctx context.Context, staleAfter time.Duration) (string, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	staleBefore := now.Add(-staleAfter)

	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return "", false, err
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback(ctx)
	}()

	q := sqlc.New(w.pool).WithTx(tx)
	uploadID, err := q.ClaimNextTournamentImport(ctx, sqlc.ClaimNextTournamentImportParams{
		Now:         pgtype.Timestamptz{Time: now, Valid: true},
		StaleBefore: pgtype.Timestamptz{Time: staleBefore, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", false, err
	}
	committed = true

	return uploadID, true, nil
}

func (w *TournamentImportWorker) processTournamentImport(ctx context.Context, uploadID string) {
	if uploadID == "" {
		return
	}

	err := func() error {
		q := sqlc.New(w.pool)
		zipBytes, err := q.GetTournamentImportArchive(ctx, uploadID)
		if err != nil {
			return err
		}

		tmpDir, err := os.MkdirTemp("", "calcutta-tournament-import-job-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		if err := archive.UnzipToDir(zipBytes, tmpDir); err != nil {
			return err
		}

		impReport, err := importer.ImportFromDir(ctx, w.pool, tmpDir, importer.Options{DryRun: false})
		if err != nil {
			return err
		}

		verReport, err := verifier.VerifyDirAgainstDB(ctx, w.pool, tmpDir)
		if err != nil {
			return err
		}

		impJSON, err := json.Marshal(impReport)
		if err != nil {
			return fmt.Errorf("marshal import report: %w", err)
		}
		verJSON, err := json.Marshal(verReport)
		if err != nil {
			return fmt.Errorf("marshal verify report: %w", err)
		}

		err = q.MarkTournamentImportSucceeded(ctx, sqlc.MarkTournamentImportSucceededParams{
			UploadID:     uploadID,
			ImportReport: impJSON,
			VerifyReport: verJSON,
		})
		if err != nil {
			slog.Error("tournament_import_mark_succeeded_failed", "upload_id", uploadID, "error", err)
		}
		return nil
	}()
	if err != nil {
		w.failTournamentImport(ctx, uploadID, err)
	}
}

func (w *TournamentImportWorker) failTournamentImport(ctx context.Context, uploadID string, jobErr error) {
	if uploadID == "" || jobErr == nil {
		return
	}

	msg := jobErr.Error()
	if errors.Is(jobErr, pgx.ErrNoRows) {
		msg = "tournament import not found"
	}

	q := sqlc.New(w.pool)
	err := q.MarkTournamentImportFailed(ctx, sqlc.MarkTournamentImportFailedParams{
		UploadID:     uploadID,
		ErrorMessage: &msg,
	})
	if err != nil {
		slog.Error("tournament_import_mark_failed_failed", "upload_id", uploadID, "error", err)
	}
}
