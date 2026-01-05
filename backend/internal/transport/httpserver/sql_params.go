package httpserver

// Helpers for building SQL query params.

func nullUUIDParam(v string) any {
	if v == "" {
		return nil
	}
	return v
}
