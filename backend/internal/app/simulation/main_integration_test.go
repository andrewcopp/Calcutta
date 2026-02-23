//go:build integration

package simulation

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	p, cleanup, err := testutil.StartPostgresContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}
	pool = p
	code := m.Run()
	cleanup()
	os.Exit(code)
}
