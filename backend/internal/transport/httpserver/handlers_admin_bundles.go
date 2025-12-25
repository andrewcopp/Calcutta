package httpserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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
	UploadID     string          `json:"upload_id"`
	Filename     string          `json:"filename"`
	SHA256       string          `json:"sha256"`
	SizeBytes    int             `json:"size_bytes"`
	ImportReport importer.Report `json:"import_report"`
	VerifyReport verifier.Report `json:"verify_report"`
}

func (s *Server) registerAdminBundleRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/bundles/export", s.adminBundlesExportHandler).Methods("GET")
	r.HandleFunc("/api/admin/bundles/import", s.adminBundlesImportHandler).Methods("POST")
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

	tmpDir, err := os.MkdirTemp("", "calcutta-bundles-import-*")
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	if err := archive.UnzipToDir(zipBytes, tmpDir); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_request", "Invalid bundle archive", "")
		return
	}

	impReport, err := importer.ImportFromDir(r.Context(), s.pool, tmpDir, importer.Options{DryRun: false})
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	verReport, err := verifier.VerifyDirAgainstDB(r.Context(), s.pool, tmpDir)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	impJSON, _ := json.Marshal(impReport)
	verJSON, _ := json.Marshal(verReport)

	var uploadID string
	err = s.pool.QueryRow(r.Context(), `
		INSERT INTO bundle_uploads (filename, sha256, size_bytes, archive, import_report, verify_report)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, header.Filename, sha, len(zipBytes), zipBytes, impJSON, verJSON).Scan(&uploadID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, adminBundlesImportResponse{UploadID: uploadID, Filename: header.Filename, SHA256: sha, SizeBytes: len(zipBytes), ImportReport: impReport, VerifyReport: verReport})
}
