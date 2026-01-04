package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/gorilla/mux"
)

func Run() {
	slog.Info("server_initializing")

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		slog.Error("config_load_failed", "error", err)
		os.Exit(1)
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, &platform.PGXPoolOptions{
		MaxConns:          cfg.PGXPoolMaxConns,
		MinConns:          cfg.PGXPoolMinConns,
		MaxConnLifetime:   time.Duration(cfg.PGXPoolMaxConnLifetimeSeconds) * time.Second,
		HealthCheckPeriod: time.Duration(cfg.PGXPoolHealthCheckPeriodSeconds) * time.Second,
	})
	if err != nil {
		slog.Error("db_connect_failed", "error", err)
		os.Exit(1)
	}

	server := NewServer(pool, cfg)

	// Router
	r := mux.NewRouter()
	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware(cfg.AllowedOrigins))
	r.Use(rateLimitMiddleware(cfg.RateLimitRPM))
	r.Use(maxBodyBytesMiddleware(cfg.HTTPMaxBodyBytes))
	r.Use(server.authenticateMiddleware)

	// Routes
	server.RegisterRoutes(r)

	// Not Found handler
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger(r.Context()).WarnContext(
			r.Context(),
			"http_not_found",
			"event", "http_not_found",
			"method", r.Method,
			"path", r.URL.Path,
		)
		writeError(w, r, http.StatusNotFound, "not_found", "Not Found", "")
	})

	port := cfg.Port

	// HTTP server with production timeouts
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadTimeout:       time.Duration(cfg.HTTPReadTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(cfg.HTTPWriteTimeoutSeconds) * time.Second,
		IdleTimeout:       time.Duration(cfg.HTTPIdleTimeoutSeconds) * time.Second,
		ReadHeaderTimeout: time.Duration(cfg.HTTPReadHeaderTimeoutSeconds) * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("server_listening", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("server_shutting_down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		slog.Error("server_shutdown_failed", "error", err)
	}

	pool.Close()

	slog.Info("server_stopped")
}
