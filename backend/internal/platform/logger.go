package platform

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strings"
)

type slogWriter struct {
	logger *slog.Logger
	level  slog.Level
}

func (w *slogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	if msg == "" {
		return len(p), nil
	}
	w.logger.Log(context.Background(), w.level, msg)
	return len(p), nil
}

func InitLogger() {
	env := strings.TrimSpace(os.Getenv("NODE_ENV"))
	if env == "" {
		env = "development"
	}

	var handler slog.Handler
	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	log.SetFlags(0)
	log.SetOutput(&slogWriter{logger: logger, level: slog.LevelInfo})
}
