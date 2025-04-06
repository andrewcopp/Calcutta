package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

func main() {
	// Parse command line flags
	calcuttaID := flag.String("calcutta", "", "Calcutta ID")
	flag.Parse()

	if *calcuttaID == "" {
		log.Fatal("Calcutta ID is required")
	}

	// Get database connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	calcuttaRepo := services.NewCalcuttaRepository(db)

	// Get the Calcutta
	calcutta, err := calcuttaRepo.GetByID(context.Background(), *calcuttaID)
	if err != nil {
		log.Fatalf("Error getting Calcutta: %v", err)
	}

	// Get all entries for the Calcutta
	entries, err := calcuttaRepo.GetEntries(context.Background(), calcutta.ID)
	if err != nil {
		log.Fatalf("Error getting entries: %v", err)
	}

	// Create a portfolio for each entry
	for _, entry := range entries {
		// Check if portfolio already exists
		portfolios, err := calcuttaRepo.GetPortfoliosByEntry(context.Background(), entry.ID)
		if err != nil {
			log.Printf("Error checking portfolios for entry %s: %v", entry.ID, err)
			continue
		}

		// If portfolio already exists, skip
		if len(portfolios) > 0 {
			log.Printf("Portfolio already exists for entry %s, skipping", entry.ID)
			continue
		}

		// Create a new portfolio
		portfolio := &models.CalcuttaPortfolio{
			EntryID: entry.ID,
		}

		err = calcuttaRepo.CreatePortfolio(context.Background(), portfolio)
		if err != nil {
			log.Printf("Error creating portfolio for entry %s: %v", entry.ID, err)
			continue
		}

		log.Printf("Created portfolio %s for entry %s", portfolio.ID, entry.ID)

		// Get entry teams
		entryTeams, err := calcuttaRepo.GetEntryTeams(context.Background(), entry.ID)
		if err != nil {
			log.Printf("Error getting entry teams for entry %s: %v", entry.ID, err)
			continue
		}

		// Create portfolio teams
		now := time.Now()
		for _, entryTeam := range entryTeams {
			// Create a new portfolio team
			portfolioTeam := &models.CalcuttaPortfolioTeam{
				PortfolioID: portfolio.ID,
				TeamID:      entryTeam.TeamID,
				Created:     now,
				Updated:     now,
			}

			err = calcuttaRepo.CreatePortfolioTeam(context.Background(), portfolioTeam)
			if err != nil {
				log.Printf("Error creating portfolio team for team %s: %v", entryTeam.TeamID, err)
				continue
			}

			log.Printf("Created portfolio team %s for team %s", portfolioTeam.ID, entryTeam.TeamID)
		}
	}

	log.Printf("All portfolios for Calcutta %s have been created", calcutta.ID)
}
