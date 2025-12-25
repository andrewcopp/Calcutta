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

func (s *Server) StartBundleImportWorker(ctx context.Context) {
	if s.pool == nil {
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case uploadID := <-s.bundleImportQueue:
				s.processBundleUpload(ctx, uploadID)
			}
		}
	}()
}

func (s *Server) EnqueuePendingBundleUploads(ctx context.Context) {
	if s.pool == nil {
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id
		FROM bundle_uploads
		WHERE deleted_at IS NULL AND status IN ('pending', 'running')
		ORDER BY created_at ASC
		LIMIT 32
	`)
	if err != nil {
		log.Printf("Error scanning pending bundle uploads: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning bundle upload id: %v", err)
			continue
		}

		select {
		case s.bundleImportQueue <- id:
			// queued
		default:
			return
		}
	}
}

func (s *Server) processBundleUpload(ctx context.Context, uploadID string) {
	if uploadID == "" {
		return
	}

	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, `
		UPDATE bundle_uploads
		SET status = 'running', started_at = $2, finished_at = NULL, error_message = NULL,
		    import_report = NULL, verify_report = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, uploadID, now)
	if err != nil {
		log.Printf("Error marking bundle upload %s running: %v", uploadID, err)
		return
	}

	var zipBytes []byte
	err = s.pool.QueryRow(ctx, `
		SELECT archive
		FROM bundle_uploads
		WHERE id = $1 AND deleted_at IS NULL
	`, uploadID).Scan(&zipBytes)
	if err != nil {
		s.failBundleUpload(ctx, uploadID, err)
		return
	}

	tmpDir, err := os.MkdirTemp("", "calcutta-bundles-import-job-*")
	if err != nil {
		s.failBundleUpload(ctx, uploadID, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	if err := archive.UnzipToDir(zipBytes, tmpDir); err != nil {
		s.failBundleUpload(ctx, uploadID, err)
		return
	}

	impReport, err := importer.ImportFromDir(ctx, s.pool, tmpDir, importer.Options{DryRun: false})
	if err != nil {
		s.failBundleUpload(ctx, uploadID, err)
		return
	}

	verReport, err := verifier.VerifyDirAgainstDB(ctx, s.pool, tmpDir)
	if err != nil {
		s.failBundleUpload(ctx, uploadID, err)
		return
	}

	calcuttas, err := s.calcuttaService.GetAllCalcuttas(ctx)
	if err != nil {
		s.failBundleUpload(ctx, uploadID, err)
		return
	}
	for _, c := range calcuttas {
		if err := s.calcuttaService.EnsurePortfoliosAndRecalculate(ctx, c.ID); err != nil {
			log.Printf("Error ensuring portfolios/recalculating for calcutta %s (upload %s): %v", c.ID, uploadID, err)
		}
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
