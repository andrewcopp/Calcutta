package httputil

import (
	"net/http"
	"strconv"
)

// GetQueryInt reads an integer query parameter from the request, returning
// defaultValue when the parameter is missing or not a valid integer.
func GetQueryInt(r *http.Request, key string, defaultValue int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return v
}
