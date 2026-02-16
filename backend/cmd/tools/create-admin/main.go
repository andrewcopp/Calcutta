package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/google/uuid"
)

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("create_admin_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	email := flag.String("email", "", "Admin user email (required)")
	name := flag.String("name", "", "Admin user name (optional, defaults to email)")
	dryRun := flag.Bool("dry-run", false, "Print SQL without executing")
	flag.Parse()

	if *email == "" {
		return fmt.Errorf("email is required (use -email=admin@example.com)")
	}

	// Validate email format (basic check)
	if !strings.Contains(*email, "@") || !strings.Contains(*email, ".") {
		return fmt.Errorf("invalid email format: %s", *email)
	}

	// Default name to email prefix if not provided
	displayName := *name
	if displayName == "" {
		displayName = strings.Split(*email, "@")[0]
	}

	// Generate SQL
	userID := uuid.New().String()
	sql := fmt.Sprintf(`
		INSERT INTO core.users (id, email, name, role, status, created_at, updated_at)
		VALUES (
			'%s',
			'%s',
			'%s',
			'admin',
			'active',
			NOW(),
			NOW()
		)
		ON CONFLICT (email) DO UPDATE
		SET
			role = 'admin',
			status = 'active',
			updated_at = NOW();
	`, userID, *email, displayName)

	if *dryRun {
		fmt.Println("Dry run mode - SQL that would be executed:")
		fmt.Println(sql)
		return nil
	}

	// Connect to database
	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Execute SQL
	tag, err := pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		slog.Info("admin_user_already_exists", "email", *email)
		fmt.Printf("Admin user already exists: %s\n", *email)
	} else {
		slog.Info("admin_user_created", "email", *email, "id", userID)
		fmt.Printf("Admin user created successfully:\n")
		fmt.Printf("  ID:     %s\n", userID)
		fmt.Printf("  Email:  %s\n", *email)
		fmt.Printf("  Name:   %s\n", displayName)
		fmt.Printf("  Role:   admin\n")
		fmt.Printf("  Status: active\n")
	}

	return nil
}
