package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/middleware"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/requestctx"
	"github.com/gorilla/mux"
)

// version is set at build time via -ldflags.
var version = "dev"

func Run() error {
	slog.Info("server_initializing")

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("config_load_failed: %w", err)
	}

	sentryCleanup, err := platform.InitSentry(cfg.SentryDSN, cfg.SentryEnvironment, version)
	if err != nil {
		return fmt.Errorf("sentry_init_failed: %w", err)
	}
	defer sentryCleanup()

	pool, err := platform.OpenPGXPool(context.Background(), cfg, &platform.PGXPoolOptions{
		MaxConns:          cfg.PGXPoolMaxConns,
		MinConns:          cfg.PGXPoolMinConns,
		MaxConnLifetime:   time.Duration(cfg.PGXPoolMaxConnLifetimeSeconds) * time.Second,
		HealthCheckPeriod: time.Duration(cfg.PGXPoolHealthCheckPeriodSeconds) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("db_connect_failed: %w", err)
	}
	defer pool.Close()

	server, err := NewServer(pool, cfg)
	if err != nil {
		return fmt.Errorf("server_init_failed: %w", err)
	}
	// Admin bootstrap is now handled by the migrate command with -bootstrap flag

	// Router
	r := mux.NewRouter()
	r.Use(requestIDMiddleware)
	r.Use(middleware.SentryMiddleware)
	r.Use(middleware.SecurityHeadersMiddleware)
	r.Use(server.loggingMiddleware)
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	r.Use(server.rateLimitMiddleware(cfg.RateLimitRPM))
	r.Use(middleware.MaxBodyBytesMiddleware(cfg.HTTPMaxBodyBytes))
	r.Use(server.authenticateMiddleware)

	// Routes
	server.RegisterRoutes(r)

	// Not Found handler
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestctx.Logger(r.Context()).WarnContext(
			r.Context(),
			"http_not_found",
			"event", "http_not_found",
			"method", r.Method,
			"path", r.URL.Path,
		)
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Not Found", "")
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	// Start server in goroutine
	go func() {
		slog.Info("server_listening", "port", port)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("server_shutting_down")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server_failed: %w", err)
		}
		slog.Info("server_shutting_down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server_shutdown_failed: %w", err)
	}

	slog.Info("server_stopped")
	return nil
}
