package main

import (
	"log/slog"
	"os"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver"
)

func main() {
	platform.InitLogger()
	if err := httpserver.Run(); err != nil {
		slog.Error("server_run_failed", "error", err)
		os.Exit(1)
	}
}
