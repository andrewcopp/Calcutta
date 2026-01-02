package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultBundleWorkerPollInterval = 2 * time.Second
	defaultBundleWorkerStaleAfter   = 30 * time.Minute
)

func (s *Server) RunBundleImportWorker(ctx context.Context) {
	s.RunBundleImportWorkerWithOptions(ctx, defaultBundleWorkerPollInterval, defaultBundleWorkerStaleAfter)
}

func (s *Server) RunBundleImportWorkerWithOptions(ctx context.Context, pollInterval time.Duration, staleAfter time.Duration) {
	if s.pool == nil {
		log.Printf("Bundle import worker disabled: database pool not available")
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
			uploadID, ok, err := s.claimNextBundleUpload(ctx, staleAfter)
			if err != nil {
				log.Printf("Error claiming next bundle upload: %v", err)
				continue
			}
			if !ok {
				continue
			}

			s.processBundleUpload(ctx, uploadID)
		}
	}
}

func (s *Server) claimNextBundleUpload(ctx context.Context, staleAfter time.Duration) (string, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	now := time.Now().UTC()
	staleBefore := now.Add(-staleAfter)

	tx, err := s.pool.Begin(ctx)
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

	q := sqlc.New(s.pool).WithTx(tx)
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

func (s *Server) processBundleUpload(ctx context.Context, uploadID string) {
	if uploadID == "" {
		return
	}

	err := func() error {
		q := sqlc.New(s.pool)
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

		impReport, err := importer.ImportFromDir(ctx, s.pool, tmpDir, importer.Options{DryRun: false})
		if err != nil {
			return err
		}

		verReport, err := verifier.VerifyDirAgainstDB(ctx, s.pool, tmpDir)
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
			log.Printf("Error marking bundle upload %s succeeded: %v", uploadID, err)
		}
		return nil
	}()
	if err != nil {
		s.failBundleUpload(ctx, uploadID, err)
	}
}

func (s *Server) failBundleUpload(ctx context.Context, uploadID string, jobErr error) {
	if uploadID == "" || jobErr == nil {
		return
	}

	msg := jobErr.Error()
	if errors.Is(jobErr, pgx.ErrNoRows) {
		msg = "bundle upload not found"
	}

	q := sqlc.New(s.pool)
	err := q.MarkBundleUploadFailed(ctx, sqlc.MarkBundleUploadFailedParams{
		UploadID:     uploadID,
		ErrorMessage: &msg,
	})
	if err != nil {
		log.Printf("Error marking bundle upload %s failed: %v", uploadID, err)
	}
}
