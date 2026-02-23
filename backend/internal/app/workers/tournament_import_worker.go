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
	"github.com/andrewcopp/Calcutta/backend/internal/app/jobqueue"
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
	pool     *pgxpool.Pool
	enqueuer *jobqueue.Enqueuer
}

func NewTournamentImportWorker(pool *pgxpool.Pool) *TournamentImportWorker {
	return &TournamentImportWorker{
		pool:     pool,
		enqueuer: jobqueue.NewEnqueuer(pool),
	}
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

	report, err := func() (*importer.Report, error) {
		q := sqlc.New(w.pool)
		zipBytes, err := q.GetTournamentImportArchive(ctx, uploadID)
		if err != nil {
			return nil, err
		}

		tmpDir, err := os.MkdirTemp("", "calcutta-tournament-import-job-*")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tmpDir)

		if err := archive.UnzipToDir(zipBytes, tmpDir); err != nil {
			return nil, err
		}

		impReport, err := importer.ImportFromDir(ctx, w.pool, tmpDir, importer.Options{DryRun: false})
		if err != nil {
			return nil, err
		}

		verReport, err := verifier.VerifyDirAgainstDB(ctx, w.pool, tmpDir)
		if err != nil {
			return nil, err
		}

		impJSON, err := json.Marshal(impReport)
		if err != nil {
			return nil, fmt.Errorf("marshal import report: %w", err)
		}
		verJSON, err := json.Marshal(verReport)
		if err != nil {
			return nil, fmt.Errorf("marshal verify report: %w", err)
		}

		err = q.MarkTournamentImportSucceeded(ctx, sqlc.MarkTournamentImportSucceededParams{
			UploadID:     uploadID,
			ImportReport: impJSON,
			VerifyReport: verJSON,
		})
		if err != nil {
			slog.Error("tournament_import_mark_succeeded_failed", "upload_id", uploadID, "error", err)
		}
		return &impReport, nil
	}()
	if err != nil {
		w.failTournamentImport(ctx, uploadID, err)
		return
	}

	w.refreshPredictions(ctx, report.TournamentIDs)
}

func (w *TournamentImportWorker) refreshPredictions(ctx context.Context, tournamentIDs []string) {
	for _, tid := range tournamentIDs {
		params, _ := json.Marshal(map[string]string{
			"tournamentId":         tid,
			"probabilitySourceKey": "kenpom",
		})
		dedupKey := fmt.Sprintf("prediction:%s", tid)
		result, err := w.enqueuer.Enqueue(ctx, jobqueue.KindRefreshPredictions, params, jobqueue.PriorityCoreApp, dedupKey)
		if err != nil {
			slog.Warn("prediction_enqueue_failed", "tournament_id", tid, "error", err)
			continue
		}
		if result.Enqueued {
			slog.Info("prediction_enqueued", "tournament_id", tid, "job_id", result.JobID)
		} else {
			slog.Info("prediction_deduplicated", "tournament_id", tid)
		}
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
