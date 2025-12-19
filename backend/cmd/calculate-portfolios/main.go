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

	// Initialize repositories and services
	calcuttaRepo := services.NewCalcuttaRepository(db)
	tournamentRepo := services.NewTournamentRepository(db)
	calcuttaService := services.NewCalcuttaService(services.CalcuttaServicePorts{
		CalcuttaReader:  calcuttaRepo,
		CalcuttaWriter:  calcuttaRepo,
		EntryReader:     calcuttaRepo,
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

	// Get the tournament
	tournament, err := tournamentRepo.GetByID(context.Background(), calcutta.TournamentID)
	if err != nil {
		log.Fatalf("Error getting tournament: %v", err)
	}

	// Get all teams for the tournament
	teams, err := tournamentRepo.GetTeams(context.Background(), tournament.ID)
	if err != nil {
		log.Fatalf("Error getting teams: %v", err)
	}

	// Get all rounds for the Calcutta
	rounds, err := calcuttaRepo.GetRounds(context.Background(), calcutta.ID)
	if err != nil {
		log.Fatalf("Error getting rounds: %v", err)
	}

	// Get all entries for the Calcutta
	entries, err := calcuttaRepo.GetEntries(context.Background(), calcutta.ID)
	if err != nil {
		log.Fatalf("Error getting entries: %v", err)
	}

	// Get all entry teams for the Calcutta
	var allEntryTeams []*models.CalcuttaEntryTeam
	for _, entry := range entries {
		entryTeams, err := calcuttaRepo.GetEntryTeams(context.Background(), entry.ID)
		if err != nil {
			log.Printf("Error getting entry teams for entry %s: %v", entry.ID, err)
			continue
		}
		allEntryTeams = append(allEntryTeams, entryTeams...)
	}

	// Calculate points for each team
	teamPoints := make(map[string]float64)
	for _, team := range teams {
		points := calcuttaService.CalculatePoints(team, rounds)
		teamPoints[team.ID] = points
	}

	// Process each entry
	for _, entry := range entries {
		// Get portfolios for this entry
		portfolios, err := calcuttaRepo.GetPortfoliosByEntry(context.Background(), entry.ID)
		if err != nil {
			log.Printf("Error getting portfolios for entry %s: %v", entry.ID, err)
			continue
		}

		// If no portfolios exist, create one
		if len(portfolios) == 0 {
			log.Printf("No portfolio exists for entry %s, creating one", entry.ID)

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

			// Add the newly created portfolio to the portfolios slice
			portfolios = append(portfolios, portfolio)
		}

		// Process each portfolio
		for _, portfolio := range portfolios {
			// Get portfolio teams
			portfolioTeams, err := calcuttaRepo.GetPortfolioTeams(context.Background(), portfolio.ID)
			if err != nil {
				log.Printf("Error getting portfolio teams for portfolio %s: %v", portfolio.ID, err)
				continue
			}

			// Update each portfolio team
			now := time.Now()
			for _, portfolioTeam := range portfolioTeams {
				// Calculate ownership percentage
				ownershipPercentage := calcuttaService.CalculateOwnershipPercentage(context.Background(), portfolioTeam, allEntryTeams)

				// Calculate points earned
				actualPoints := 0.0
				if points, ok := teamPoints[portfolioTeam.TeamID]; ok {
					actualPoints = points * ownershipPercentage
				}

				// Update portfolio team
				portfolioTeam.OwnershipPercentage = ownershipPercentage
				portfolioTeam.ActualPoints = actualPoints
				portfolioTeam.Updated = now

				err = calcuttaRepo.UpdatePortfolioTeam(context.Background(), portfolioTeam)
				if err != nil {
					log.Printf("Error updating portfolio team %s: %v", portfolioTeam.ID, err)
					continue
				}
			}

			// Update the portfolio's updated timestamp
			portfolio.Updated = now
			err = calcuttaRepo.UpdatePortfolio(context.Background(), portfolio)
			if err != nil {
				log.Printf("Error updating portfolio %s: %v", portfolio.ID, err)
				continue
			}

			// Calculate total points for the portfolio
			totalPoints := calcuttaService.CalculatePlayerPoints(portfolio, portfolioTeams)
			log.Printf("Portfolio %s updated with total points: %.2f", portfolio.ID, totalPoints)
		}
	}

	log.Printf("All portfolios for Calcutta %s have been calculated", calcutta.ID)
}
