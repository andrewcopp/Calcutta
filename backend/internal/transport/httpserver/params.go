package httpserver

import (
	"net/http"
	"strconv"
)

func getQueryInt(r *http.Request, key string, defaultValue int) int {
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

func getLimit(r *http.Request, defaultValue int) int {
	return getQueryInt(r, "limit", defaultValue)
}

func getOffset(r *http.Request, defaultValue int) int {
	return getQueryInt(r, "offset", defaultValue)
}
