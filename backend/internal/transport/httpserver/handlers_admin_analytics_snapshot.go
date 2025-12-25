package httpserver

import (
	"net/http"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/analytics_snapshot"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/gorilla/mux"
)

func (s *Server) registerAdminAnalyticsRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/analytics/export", s.requirePermission("admin.analytics.export", s.adminAnalyticsExportHandler)).Methods("GET")
}

func (s *Server) adminAnalyticsExportHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	tournamentID := r.URL.Query().Get("tournamentId")
	calcuttaID := r.URL.Query().Get("calcuttaId")
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "tournamentId is required", "tournamentId")
		return
	}
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "calcuttaId is required", "calcuttaId")
		return
	}

	bracket, err := s.app.Bracket.GetBracket(r.Context(), tournamentID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	tmpDir, err := os.MkdirTemp("", "calcutta-analytics-snapshot-*")
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer os.RemoveAll(tmpDir)

	generatedAt := time.Now().UTC()
	res, err := analytics_snapshot.ExportToDir(
		r.Context(),
		s.pool,
		tmpDir,
		generatedAt,
		tournamentID,
		calcuttaID,
		analytics_snapshot.ExportInputs{Bracket: bracket},
	)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	zipBytes, err := archive.ZipDir(tmpDir)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	filename := "analytics-" + res.Manifest.TournamentKey + "-" + res.Manifest.CalcuttaKey + "-" + generatedAt.Format("20060102-150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(zipBytes)
}
