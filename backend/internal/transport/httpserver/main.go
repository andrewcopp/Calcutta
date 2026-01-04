package httpserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/gorilla/mux"
)

func Run() {
	log.Printf("Starting server initialization...")

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := platform.OpenPGXPool(context.Background(), cfg, &platform.PGXPoolOptions{
		MaxConns:          cfg.PGXPoolMaxConns,
		MinConns:          cfg.PGXPoolMinConns,
		MaxConnLifetime:   time.Duration(cfg.PGXPoolMaxConnLifetimeSeconds) * time.Second,
		HealthCheckPeriod: time.Duration(cfg.PGXPoolHealthCheckPeriodSeconds) * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database (pgxpool): %v", err)
	}

	server := NewServer(pool, cfg)

	// Router
	r := mux.NewRouter()
	r.Use(requestIDMiddleware)
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)
	r.Use(maxBodyBytesMiddleware(cfg.HTTPMaxBodyBytes))
	r.Use(server.authenticateMiddleware)

	// Routes
	server.RegisterRoutes(r)

	// Not Found handler
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] No route matched: %s %s", getRequestID(r.Context()), r.Method, r.URL.Path)
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
		log.Printf("Server starting on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	pool.Close()

	log.Println("Server stopped gracefully")
}
