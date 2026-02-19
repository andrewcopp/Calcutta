package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line flags
	up := flag.Bool("up", false, "Run migrations up")
	down := flag.Bool("down", false, "Run migrations down")
	force := flag.Int("force", 0, "Force schema migration version (clears dirty state)")
	bootstrap := flag.Bool("bootstrap", false, "Bootstrap admin user after migrations")
	flag.Parse()

	// Check if at least one flag is set
	if !*up && !*down && *force == 0 && !*bootstrap {
		fmt.Println("Please specify either -up, -down, or -bootstrap flag")
		flag.Usage()
		return fmt.Errorf("no migration action specified")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	m, err := newMigrator(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Run migrations
	if *force != 0 {
		fmt.Printf("Forcing schema migration version to %d (clearing dirty state)...\n", *force)
		if err := m.Force(*force); err != nil {
			return fmt.Errorf("error forcing schema migrations: %w", err)
		}
		fmt.Println("Schema migration version forced successfully")
	}

	if *up {
		fmt.Println("Running schema migrations up...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running schema migrations: %w", err)
		}
		fmt.Println("Schema migrations completed successfully")
	}

	if *down {
		fmt.Println("Rolling back schema migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error rolling back schema migrations: %w", err)
		}
		fmt.Println("Schema migrations rolled back successfully")
	}

	if *bootstrap {
		if err := bootstrapAdmin(cfg); err != nil {
			return fmt.Errorf("error bootstrapping admin: %w", err)
		}
	}

	return nil
}

func bootstrapAdmin(cfg platform.Config) error {
	email := strings.TrimSpace(cfg.BootstrapAdminEmail)
	if email == "" {
		fmt.Println("No BOOTSTRAP_ADMIN_EMAIL set, skipping admin bootstrap")
		return nil
	}

	password := strings.TrimSpace(cfg.BootstrapAdminPassword)
	if password == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_PASSWORD must be set when BOOTSTRAP_ADMIN_EMAIL is set")
	}

	fmt.Printf("Bootstrapping admin user: %s\n", email)

	ctx := context.Background()
	pool, err := platform.OpenPGXPool(ctx, cfg, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	userRepo := dbadapters.NewUserRepository(pool)
	authzRepo := dbadapters.NewAuthorizationRepository(pool)

	user, err := userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	if user == nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		h := string(hash)
		user = &models.User{
			ID:           uuid.New().String(),
			Email:        &email,
			FirstName:    "Admin",
			LastName:     "User",
			Status:       "active",
			PasswordHash: &h,
		}
		if err := userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		fmt.Println("Created admin user")
	} else {
		fmt.Println("Admin user already exists")
	}

	if err := authzRepo.GrantGlobalAdmin(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to grant admin permissions: %w", err)
	}
	fmt.Println("Admin bootstrap completed successfully")

	return nil
}

func newMigrator(databaseURL string) (*migrate.Migrate, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting working directory: %w", err)
	}

	migrationsDir := filepath.Join(workDir, "migrations", "schema")
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error creating migrator: %w", err)
	}

	return m, nil
}
