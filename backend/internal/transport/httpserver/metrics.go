package httpserver

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

var httpMetrics = newHTTPMetrics()

type httpMetricsRegistry struct {
	mu sync.Mutex

	requestsTotal    int64
	requestsByMethod map[string]int64
	requestsByStatus map[int]int64

	latencyCount   int64
	latencySumMs   int64
	latencyMaxMs   int64
	latencyBuckets map[int64]int64
}

func newHTTPMetrics() *httpMetricsRegistry {
	return &httpMetricsRegistry{
		requestsByMethod: map[string]int64{},
		requestsByStatus: map[int]int64{},
		latencyBuckets: map[int64]int64{
			10:   0,
			25:   0,
			50:   0,
			100:  0,
			250:  0,
			500:  0,
			1000: 0,
			-1:   0,
		},
	}
}

func (m *httpMetricsRegistry) Observe(method string, status int, duration time.Duration) {
	durMs := duration.Milliseconds()
	if durMs < 0 {
		durMs = 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestsTotal++
	m.requestsByMethod[method]++
	m.requestsByStatus[status]++

	m.latencyCount++
	m.latencySumMs += durMs
	if durMs > m.latencyMaxMs {
		m.latencyMaxMs = durMs
	}
	for _, b := range []int64{10, 25, 50, 100, 250, 500, 1000} {
		if durMs <= b {
			m.latencyBuckets[b]++
			return
		}
	}
	m.latencyBuckets[-1]++
}

func (m *httpMetricsRegistry) WritePrometheus(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_http_requests_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_http_requests_total %d\n", m.requestsTotal)

	methods := make([]string, 0, len(m.requestsByMethod))
	for k := range m.requestsByMethod {
		methods = append(methods, k)
	}
	sort.Strings(methods)
	_, _ = fmt.Fprintln(w, "# TYPE calcutta_http_requests_by_method_total counter")
	for _, method := range methods {
		_, _ = fmt.Fprintf(w, "calcutta_http_requests_by_method_total{method=%q} %d\n", method, m.requestsByMethod[method])
	}

	statuses := make([]int, 0, len(m.requestsByStatus))
	for s := range m.requestsByStatus {
		statuses = append(statuses, s)
	}
	sort.Ints(statuses)
	_, _ = fmt.Fprintln(w, "# TYPE calcutta_http_requests_by_status_total counter")
	for _, status := range statuses {
		_, _ = fmt.Fprintf(w, "calcutta_http_requests_by_status_total{status=%q} %d\n", strconv.Itoa(status), m.requestsByStatus[status])
	}

	avg := float64(0)
	if m.latencyCount > 0 {
		avg = float64(m.latencySumMs) / float64(m.latencyCount)
	}

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_http_request_duration_ms_avg gauge")
	_, _ = fmt.Fprintf(w, "calcutta_http_request_duration_ms_avg %.3f\n", avg)
	_, _ = fmt.Fprintln(w, "# TYPE calcutta_http_request_duration_ms_max gauge")
	_, _ = fmt.Fprintf(w, "calcutta_http_request_duration_ms_max %d\n", m.latencyMaxMs)

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_http_request_duration_ms_bucket counter")
	for _, b := range []int64{10, 25, 50, 100, 250, 500, 1000, -1} {
		label := "inf"
		if b >= 0 {
			label = strconv.FormatInt(b, 10)
		}
		_, _ = fmt.Fprintf(w, "calcutta_http_request_duration_ms_bucket{le=%q} %d\n", label, m.latencyBuckets[b])
	}
}

func writePGXPoolPrometheus(w io.Writer, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}

	st := pool.Stat()

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_max_conns gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_max_conns %d\n", st.MaxConns())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_total_conns gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_total_conns %d\n", st.TotalConns())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_idle_conns gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_idle_conns %d\n", st.IdleConns())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_acquired_conns gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_acquired_conns %d\n", st.AcquiredConns())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_constructing_conns gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_constructing_conns %d\n", st.ConstructingConns())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_acquire_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_acquire_total %d\n", st.AcquireCount())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_empty_acquire_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_empty_acquire_total %d\n", st.EmptyAcquireCount())

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_canceled_acquire_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_canceled_acquire_total %d\n", st.CanceledAcquireCount())

	dur := st.AcquireDuration()
	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_acquire_duration_ms_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_db_pool_acquire_duration_ms_total %d\n", dur.Milliseconds())

	if cnt := st.AcquireCount(); cnt > 0 {
		avgMs := float64(dur.Milliseconds()) / float64(cnt)
		_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_pool_acquire_duration_ms_avg gauge")
		_, _ = fmt.Fprintf(w, "calcutta_db_pool_acquire_duration_ms_avg %.3f\n", avgMs)
	}
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.MetricsEnabled {
		writeError(w, r, http.StatusNotFound, "not_found", "Not Found", "")
		return
	}
	if s.cfg.MetricsAuthToken != "" {
		tok := strings.TrimSpace(r.Header.Get("X-Metrics-Token"))
		if tok == "" {
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tok = strings.TrimSpace(parts[1])
			}
		}
		if tok == "" || tok != s.cfg.MetricsAuthToken {
			writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	httpMetrics.WritePrometheus(w)
	writePGXPoolPrometheus(w, s.pool)
	platform.WriteDBQueryMetricsPrometheus(w)
}
