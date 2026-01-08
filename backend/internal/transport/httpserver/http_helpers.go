package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
)

func writeJSON(w http.ResponseWriter, status int, body any) {
	response.WriteJSON(w, status, body)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string, field string) {
	httperr.Write(w, r, status, code, message, field)
}

func writeErrorFromErr(w http.ResponseWriter, r *http.Request, err error) {
	httperr.WriteFromErr(w, r, err, authUserID)
}
