package workers

import (
	"os"
	"strconv"
	"strings"
)

func resolveRunJobsMaxAttempts(defaultValue int) int {
	v := strings.TrimSpace(os.Getenv("RUN_JOBS_MAX_ATTEMPTS"))
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	if n <= 0 {
		return defaultValue
	}
	return n
}
