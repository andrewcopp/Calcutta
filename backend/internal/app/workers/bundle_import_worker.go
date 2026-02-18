package workers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultBundleWorkerPollInterval = 2 * time.Second
	defaultBundleWorkerStaleAfter   = 30 * time.Minute
)

type BundleImportWorker struct {
	pool *pgxpool.Pool
}

func NewBundleImportWorker(pool *pgxpool.Pool) *BundleImportWorker {
	return &BundleImportWorker{pool: pool}
}

func (w *BundleImportWorker) Run(ctx context.Context) {
	w.RunWithOptions(ctx, defaultBundleWorkerPollInterval, defaultBundleWorkerStaleAfter)
}

func (w *BundleImportWorker) RunWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if w == nil || w.pool == nil {
		slog.Warn("bundle_import_worker_disabled", "reason", "database pool not available")
		<-ctx.Done()
		return
	}
	if pollInterval <= 0 {
		pollInterval = defaultBundleWorkerPollInterval
	}
	if staleAfter <= 0 {
		staleAfter = defaultBundleWorkerStaleAfter
	}

	t := time.NewTicker(pollInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			uploadID, ok, err := w.claimNextBundleUpload(ctx, staleAfter)
			if err != nil {
				slog.Error("bundle_import_claim_failed", "error", err)
				continue
			}
			if !ok {
				continue
			}
			w.processBundleUpload(ctx, uploadID)
		}
	}
}

func (w *BundleImportWorker) claimNextBundleUpload(ctx context.Context, staleAfter time.Duration) (string, bool, error) {
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
	uploadID, err := q.ClaimNextBundleUpload(ctx, sqlc.ClaimNextBundleUploadParams{
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

func (w *BundleImportWorker) processBundleUpload(ctx context.Context, uploadID string) {
	if uploadID == "" {
		return
	}

	err := func() error {
		q := sqlc.New(w.pool)
		zipBytes, err := q.GetBundleUploadArchive(ctx, uploadID)
		if err != nil {
			return err
		}

		tmpDir, err := os.MkdirTemp("", "calcutta-bundles-import-job-*")
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

		impJSON, _ := json.Marshal(impReport)
		verJSON, _ := json.Marshal(verReport)

		err = q.MarkBundleUploadSucceeded(ctx, sqlc.MarkBundleUploadSucceededParams{
			UploadID:     uploadID,
			ImportReport: impJSON,
			VerifyReport: verJSON,
		})
		if err != nil {
			slog.Error("bundle_upload_mark_succeeded_failed", "upload_id", uploadID, "error", err)
		}
		return nil
	}()
	if err != nil {
		w.failBundleUpload(ctx, uploadID, err)
	}
}

func (w *BundleImportWorker) failBundleUpload(ctx context.Context, uploadID string, jobErr error) {
	if uploadID == "" || jobErr == nil {
		return
	}

	msg := jobErr.Error()
	if errors.Is(jobErr, pgx.ErrNoRows) {
		msg = "bundle upload not found"
	}

	q := sqlc.New(w.pool)
	err := q.MarkBundleUploadFailed(ctx, sqlc.MarkBundleUploadFailedParams{
		UploadID:     uploadID,
		ErrorMessage: &msg,
	})
	if err != nil {
		slog.Error("bundle_upload_mark_failed_failed", "upload_id", uploadID, "error", err)
	}
}
