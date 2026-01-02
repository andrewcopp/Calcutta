package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/jackc/pgx/v5"
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

	var uploadID string
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM bundle_uploads
		WHERE deleted_at IS NULL
		  AND finished_at IS NULL
		  AND (
			status = 'pending'
			OR (status = 'running' AND started_at IS NOT NULL AND started_at < $1)
		  )
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, staleBefore).Scan(&uploadID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE bundle_uploads
		SET status = 'running', started_at = $2, finished_at = NULL, error_message = NULL,
		    import_report = NULL, verify_report = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, uploadID, now)
	if err != nil {
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
		var zipBytes []byte
		err := s.pool.QueryRow(ctx, `
			SELECT archive
			FROM bundle_uploads
			WHERE id = $1 AND deleted_at IS NULL
		`, uploadID).Scan(&zipBytes)
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

		_, err = s.pool.Exec(ctx, `
			UPDATE bundle_uploads
			SET status = 'succeeded', finished_at = NOW(), import_report = $2, verify_report = $3, updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL
		`, uploadID, impJSON, verJSON)
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

	_, err := s.pool.Exec(ctx, `
		UPDATE bundle_uploads
		SET status = 'failed', finished_at = NOW(), error_message = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, uploadID, msg)
	if err != nil {
		log.Printf("Error marking bundle upload %s failed: %v", uploadID, err)
	}
}
