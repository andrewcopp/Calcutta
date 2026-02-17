package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httputil"
)

func getLimit(r *http.Request, defaultValue int) int {
	return httputil.GetQueryInt(r, "limit", defaultValue)
}
