package platform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

var dbQueryMetrics = newDBQueryMetricsRegistry()

type dbQueryMetricsRegistry struct {
	mu sync.Mutex

	queryCount int64
	queryErrs  int64

	latencyCount   int64
	latencySumMs   int64
	latencyMaxMs   int64
	latencyBuckets map[int64]int64
}

func newDBQueryMetricsRegistry() *dbQueryMetricsRegistry {
	return &dbQueryMetricsRegistry{
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

func (m *dbQueryMetricsRegistry) Observe(duration time.Duration, err error) {
	durMs := duration.Milliseconds()
	if durMs < 0 {
		durMs = 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.queryCount++
	if err != nil {
		m.queryErrs++
	}

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

func (m *dbQueryMetricsRegistry) WritePrometheus(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_queries_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_db_queries_total %d\n", m.queryCount)
	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_query_errors_total counter")
	_, _ = fmt.Fprintf(w, "calcutta_db_query_errors_total %d\n", m.queryErrs)

	avg := float64(0)
	if m.latencyCount > 0 {
		avg = float64(m.latencySumMs) / float64(m.latencyCount)
	}

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_query_duration_ms_avg gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_query_duration_ms_avg %.3f\n", avg)
	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_query_duration_ms_max gauge")
	_, _ = fmt.Fprintf(w, "calcutta_db_query_duration_ms_max %d\n", m.latencyMaxMs)

	_, _ = fmt.Fprintln(w, "# TYPE calcutta_db_query_duration_ms_bucket counter")
	for _, b := range []int64{10, 25, 50, 100, 250, 500, 1000, -1} {
		label := "inf"
		if b >= 0 {
			label = strconv.FormatInt(b, 10)
		}
		_, _ = fmt.Fprintf(w, "calcutta_db_query_duration_ms_bucket{le=%q} %d\n", label, m.latencyBuckets[b])
	}
}

func WriteDBQueryMetricsPrometheus(w io.Writer) {
	dbQueryMetrics.WritePrometheus(w)
}

type dbQueryTracer struct{}

type dbQueryTraceKey struct{}

func (t *dbQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, dbQueryTraceKey{}, time.Now())
}

func (t *dbQueryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value(dbQueryTraceKey{}).(time.Time)
	if !ok {
		return
	}
	err := data.Err
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	dbQueryMetrics.Observe(time.Since(start), err)
}

func newDBQueryTracer() pgx.QueryTracer {
	return &dbQueryTracer{}
}
