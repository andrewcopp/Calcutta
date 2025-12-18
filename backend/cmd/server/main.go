package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	schoolRepo        *services.SchoolRepository
	schoolService     *services.SchoolService
	tournamentRepo    *services.TournamentRepository
	tournamentService *services.TournamentService
	calcuttaRepo      *services.CalcuttaRepository
	calcuttaService   *services.CalcuttaService
	userRepo          *services.UserRepository
	userService       *services.UserService
)

func main() {
	log.Printf("Starting server initialization...")

	// Connect to database
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Printf("Successfully connected to database")

	// Initialize repositories
	schoolRepo = services.NewSchoolRepository(db)
	tournamentRepo = services.NewTournamentRepository(db)
	calcuttaRepo = services.NewCalcuttaRepository(db)
	userRepo = services.NewUserRepository(db)

	// Initialize services
	schoolService = services.NewSchoolService(schoolRepo)
	tournamentService = services.NewTournamentService(tournamentRepo)
	calcuttaService = services.NewCalcuttaService(calcuttaRepo)
	userService = services.NewUserService(userRepo)

	// Router
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	// Routes
	RegisterRoutes(r)

	// Not Found handler with logging
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("No route matched for request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	// Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
