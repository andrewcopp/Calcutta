package httpserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/exporter"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/gorilla/mux"
)

type adminBundlesImportResponse struct {
	UploadID  string `json:"upload_id"`
	Status    string `json:"status"`
	Filename  string `json:"filename"`
	SHA256    string `json:"sha256"`
	SizeBytes int    `json:"size_bytes"`
}

type adminBundlesImportStatusResponse struct {
	UploadID     string           `json:"upload_id"`
	Filename     string           `json:"filename"`
	SHA256       string           `json:"sha256"`
	SizeBytes    int              `json:"size_bytes"`
	Status       string           `json:"status"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	FinishedAt   *time.Time       `json:"finished_at,omitempty"`
	ErrorMessage *string          `json:"error_message,omitempty"`
	ImportReport *importer.Report `json:"import_report,omitempty"`
	VerifyReport *verifier.Report `json:"verify_report,omitempty"`
}

func (s *Server) registerAdminBundleRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/bundles/export", s.adminBundlesExportHandler).Methods("GET")
	r.HandleFunc("/api/admin/bundles/import", s.adminBundlesImportHandler).Methods("POST")
	r.HandleFunc("/api/admin/bundles/import/{uploadId}", s.adminBundlesImportStatusHandler).Methods("GET")
}

func (s *Server) adminBundlesExportHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	tmpDir, err := os.MkdirTemp("", "calcutta-bundles-export-*")
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	generatedAt := time.Now().UTC()
	if err := exporter.ExportToDir(r.Context(), s.pool, tmpDir, generatedAt); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	zipBytes, err := archive.ZipDir(tmpDir)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	filename := "bundles-" + generatedAt.Format("20060102-150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func (s *Server) adminBundlesImportHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid multipart form", "")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "file is required", "file")
		return
	}
	defer file.Close()

	zipBytes, err := io.ReadAll(file)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	sum := sha256.Sum256(zipBytes)
	sha := hex.EncodeToString(sum[:])

	var uploadID string
	err = s.pool.QueryRow(r.Context(), `
		INSERT INTO bundle_uploads (filename, sha256, size_bytes, archive, status)
		VALUES ($1, $2, $3, $4, 'pending')
		ON CONFLICT (sha256) WHERE deleted_at IS NULL
		DO UPDATE SET
			filename = EXCLUDED.filename,
			size_bytes = EXCLUDED.size_bytes,
			archive = EXCLUDED.archive,
			status = 'pending',
			started_at = NULL,
			finished_at = NULL,
			error_message = NULL,
			import_report = NULL,
			verify_report = NULL,
			updated_at = NOW()
		RETURNING id
	`, header.Filename, sha, len(zipBytes), zipBytes).Scan(&uploadID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	select {
	case s.bundleImportQueue <- uploadID:
		// queued
	default:
		log.Printf("Bundle import queue is full; upload %s will remain pending", uploadID)
	}

	writeJSON(w, http.StatusAccepted, adminBundlesImportResponse{UploadID: uploadID, Status: "pending", Filename: header.Filename, SHA256: sha, SizeBytes: len(zipBytes)})
}

func (s *Server) adminBundlesImportStatusHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	vars := mux.Vars(r)
	uploadID := vars["uploadId"]
	if uploadID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Upload ID is required", "uploadId")
		return
	}

	var (
		filename, sha256, status string
		sizeBytes                int
		startedAt, finishedAt    *time.Time
		errMsg                   *string
		impJSON, verJSON         []byte
	)

	err := s.pool.QueryRow(r.Context(), `
		SELECT filename, sha256, size_bytes, status, started_at, finished_at, error_message,
		       COALESCE(import_report, '{}'::jsonb), COALESCE(verify_report, '{}'::jsonb)
		FROM bundle_uploads
		WHERE id = $1 AND deleted_at IS NULL
	`, uploadID).Scan(&filename, &sha256, &sizeBytes, &status, &startedAt, &finishedAt, &errMsg, &impJSON, &verJSON)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	resp := adminBundlesImportStatusResponse{UploadID: uploadID, Filename: filename, SHA256: sha256, SizeBytes: sizeBytes, Status: status, StartedAt: startedAt, FinishedAt: finishedAt, ErrorMessage: errMsg}
	if len(impJSON) > 0 && string(impJSON) != "{}" {
		var rep importer.Report
		if err := json.Unmarshal(impJSON, &rep); err == nil {
			resp.ImportReport = &rep
		}
	}
	if len(verJSON) > 0 && string(verJSON) != "{}" {
		var rep verifier.Report
		if err := json.Unmarshal(verJSON, &rep); err == nil {
			resp.VerifyReport = &rep
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
