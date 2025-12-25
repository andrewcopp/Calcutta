package main

import (
	"context"
	"flag"
	"log"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

func main() {
	// Parse command line flags
	calcuttaID := flag.String("calcutta", "", "Calcutta ID")
	flag.Parse()

	if *calcuttaID == "" {
		log.Fatal("Calcutta ID is required")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := platform.OpenDB(context.Background(), cfg, nil)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories and services
	calcuttaRepo := services.NewCalcuttaRepository(db)
	calcuttaService := services.NewCalcuttaService(services.CalcuttaServicePorts{
		CalcuttaReader:  calcuttaRepo,
		CalcuttaWriter:  calcuttaRepo,
		EntryReader:     calcuttaRepo,
		PayoutReader:    calcuttaRepo,
		PortfolioReader: calcuttaRepo,
		PortfolioWriter: calcuttaRepo,
		RoundWriter:     calcuttaRepo,
		TeamReader:      calcuttaRepo,
	})

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

		// Create a new portfolio with correct ownership via service
		portfolio, err := calcuttaService.CreatePortfolio(context.Background(), entry.ID)
		if err != nil {
			log.Printf("Error creating portfolio for entry %s: %v", entry.ID, err)
			continue
		}
		log.Printf("Created portfolio %s for entry %s", portfolio.ID, entry.ID)
	}

	log.Printf("All portfolios for Calcutta %s have been created", calcutta.ID)
}
