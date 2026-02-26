package httpserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/exporter"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/importer"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/verifier"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

type adminTournamentImportResponse struct {
	UploadID  string `json:"uploadId"`
	Status    string `json:"status"`
	Filename  string `json:"filename"`
	SHA256    string `json:"sha256"`
	SizeBytes int    `json:"sizeBytes"`
}

type adminTournamentImportStatusResponse struct {
	UploadID     string           `json:"uploadId"`
	Filename     string           `json:"filename"`
	SHA256       string           `json:"sha256"`
	SizeBytes    int              `json:"sizeBytes"`
	Status       string           `json:"status"`
	StartedAt    *time.Time       `json:"startedAt,omitempty"`
	FinishedAt   *time.Time       `json:"finishedAt,omitempty"`
	ErrorMessage *string          `json:"errorMessage,omitempty"`
	ImportReport *importer.Report `json:"importReport,omitempty"`
	VerifyReport *verifier.Report `json:"verifyReport,omitempty"`
}

func (s *Server) registerAdminTournamentImportRoutes(r *mux.Router) {
	r.HandleFunc("/api/v1/admin/tournament-imports/export", s.requirePermission("admin.bundles.export", s.adminTournamentImportsExportHandler)).Methods("GET")
	r.HandleFunc("/api/v1/admin/tournament-imports/import", s.requirePermission("admin.bundles.import", s.adminTournamentImportsImportHandler)).Methods("POST")
	r.HandleFunc("/api/v1/admin/tournament-imports/import/{uploadId}", s.requirePermission("admin.bundles.read", s.adminTournamentImportsImportStatusHandler)).Methods("GET")
}

func (s *Server) adminTournamentImportsExportHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	tmpDir, err := os.MkdirTemp("", "calcutta-tournament-imports-export-*")
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}
	defer os.RemoveAll(tmpDir)

	generatedAt := time.Now().UTC()
	if err := exporter.ExportToDir(r.Context(), s.pool, tmpDir, generatedAt); err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	zipBytes, err := archive.ZipDir(tmpDir)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	filename := "tournament-data-" + generatedAt.Format("20060102-150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}

func (s *Server) adminTournamentImportsImportHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "invalid_request", "Invalid multipart form", "")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "file is required", "file")
		return
	}
	defer file.Close()

	zipBytes, err := io.ReadAll(file)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	sum := sha256.Sum256(zipBytes)
	sha := hex.EncodeToString(sum[:])

	q := sqlc.New(s.pool)
	uploadID, err := q.UpsertTournamentImport(r.Context(), sqlc.UpsertTournamentImportParams{
		Filename:  header.Filename,
		Sha256:    sha,
		SizeBytes: int64(len(zipBytes)),
		Archive:   zipBytes,
	})
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, adminTournamentImportResponse{UploadID: uploadID, Status: "pending", Filename: header.Filename, SHA256: sha, SizeBytes: len(zipBytes)})
}

func (s *Server) adminTournamentImportsImportStatusHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		httperr.Write(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	vars := mux.Vars(r)
	uploadID := vars["uploadId"]
	if uploadID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Upload ID is required", "uploadId")
		return
	}

	q := sqlc.New(s.pool)
	row, err := q.GetTournamentImportStatus(r.Context(), uploadID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, authUserID)
		return
	}

	var startedAt *time.Time
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		startedAt = &t
	}
	var finishedAt *time.Time
	if row.FinishedAt.Valid {
		t := row.FinishedAt.Time
		finishedAt = &t
	}

	resp := adminTournamentImportStatusResponse{UploadID: uploadID, Filename: row.Filename, SHA256: row.Sha256, SizeBytes: int(row.SizeBytes), Status: row.Status, StartedAt: startedAt, FinishedAt: finishedAt, ErrorMessage: row.ErrorMessage}
	if len(row.ImportReport) > 0 && string(row.ImportReport) != "{}" {
		var rep importer.Report
		if err := json.Unmarshal(row.ImportReport, &rep); err == nil {
			resp.ImportReport = &rep
		}
	}
	if len(row.VerifyReport) > 0 && string(row.VerifyReport) != "{}" {
		var rep verifier.Report
		if err := json.Unmarshal(row.VerifyReport, &rep); err == nil {
			resp.VerifyReport = &rep
		}
	}

	response.WriteJSON(w, http.StatusOK, resp)
}
