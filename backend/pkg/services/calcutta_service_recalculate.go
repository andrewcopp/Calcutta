package services

import (
	"context"
	"log"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// RecalculatePortfolio recalculates the portfolio for a calcutta entry
func (s *CalcuttaService) RecalculatePortfolio(ctx context.Context, calcuttaID string) error {
	log.Printf("Starting portfolio recalculation for calcutta %s", calcuttaID)

	entries, err := s.ports.EntryReader.GetEntries(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error getting entries: %v", err)
		return err
	}
	log.Printf("Found %d entries to process", len(entries))

	for _, entry := range entries {
		if err := s.recalculateEntryPortfolios(ctx, calcuttaID, entry); err != nil {
			return err
		}
	}

	log.Printf("Completed portfolio recalculation for calcutta %s", calcuttaID)
	return nil
}

func (s *CalcuttaService) recalculateEntryPortfolios(ctx context.Context, calcuttaID string, entry *models.CalcuttaEntry) error {
	log.Printf("Processing entry %s", entry.ID)

	portfolios, err := s.ports.PortfolioReader.GetPortfolios(ctx, entry.ID)
	if err != nil {
		log.Printf("Error getting portfolios for entry %s: %v", entry.ID, err)
		return err
	}
	log.Printf("Found %d portfolios for entry %s", len(portfolios), entry.ID)

	for _, portfolio := range portfolios {
		if err := s.recalculatePortfolio(ctx, calcuttaID, portfolio); err != nil {
			return err
		}
	}

	return nil
}

func (s *CalcuttaService) recalculatePortfolio(ctx context.Context, calcuttaID string, portfolio *models.CalcuttaPortfolio) error {
	log.Printf("Processing portfolio %s for entry %s", portfolio.ID, portfolio.EntryID)

	portfolioTeams, err := s.ports.PortfolioReader.GetPortfolioTeams(ctx, portfolio.ID)
	if err != nil {
		log.Printf("Error getting portfolio teams: %v", err)
		return err
	}
	log.Printf("Found %d teams in portfolio", len(portfolioTeams))

	allTeams, err := s.loadAllEntryTeams(ctx, calcuttaID)
	if err != nil {
		return err
	}

	for _, portfolioTeam := range portfolioTeams {
		if err := s.recalculatePortfolioTeam(ctx, portfolio, portfolioTeam, allTeams); err != nil {
			return err
		}
	}

	return nil
}

func (s *CalcuttaService) loadAllEntryTeams(ctx context.Context, calcuttaID string) ([]*models.CalcuttaEntryTeam, error) {
	allEntryTeams, err := s.ports.EntryReader.GetEntries(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error getting all entries: %v", err)
		return nil, err
	}

	var allTeams []*models.CalcuttaEntryTeam
	for _, entry := range allEntryTeams {
		entryTeams, err := s.ports.EntryReader.GetEntryTeams(ctx, entry.ID)
		if err != nil {
			log.Printf("Error getting entry teams for entry %s: %v", entry.ID, err)
			return nil, err
		}
		allTeams = append(allTeams, entryTeams...)
	}
	log.Printf("Found %d total entry teams across all entries", len(allTeams))

	return allTeams, nil
}

func (s *CalcuttaService) recalculatePortfolioTeam(ctx context.Context, portfolio *models.CalcuttaPortfolio, portfolioTeam *models.CalcuttaPortfolioTeam, allTeams []*models.CalcuttaEntryTeam) error {
	log.Printf("Processing team %s in portfolio", portfolioTeam.TeamID)

	tournamentTeam, err := s.ports.TeamReader.GetTournamentTeam(ctx, portfolioTeam.TeamID)
	if err != nil {
		log.Printf("Error getting tournament team: %v", err)
		return err
	}
	log.Printf("Tournament team: Wins=%d, Byes=%d, Eliminated=%v",
		tournamentTeam.Wins, tournamentTeam.Byes, tournamentTeam.Eliminated)

	teamPoints := s.CalculatePoints(tournamentTeam, nil)
	log.Printf("Calculated team points: %f", teamPoints)

	var entryTeamBid float64
	totalBidsForTeam := 0.0

	for _, et := range allTeams {
		if et.TeamID == portfolioTeam.TeamID {
			totalBidsForTeam += float64(et.Bid)
			if et.EntryID == portfolio.EntryID {
				entryTeamBid = float64(et.Bid)
				log.Printf("Found entry team bid: %f", entryTeamBid)
			}
		}
	}

	log.Printf("Total bids for team %s: %f", portfolioTeam.TeamID, totalBidsForTeam)

	ownershipPercentage := 0.0
	if totalBidsForTeam > 0 {
		ownershipPercentage = entryTeamBid / totalBidsForTeam
	}
	log.Printf("Calculated ownership percentage: %f (bid: %f / total bids: %f)",
		ownershipPercentage, entryTeamBid, totalBidsForTeam)

	oldPoints := portfolioTeam.ActualPoints
	portfolioTeam.OwnershipPercentage = ownershipPercentage
	portfolioTeam.ActualPoints = teamPoints * ownershipPercentage
	log.Printf("Updating points from %f to %f (team points %f * ownership %f)",
		oldPoints, portfolioTeam.ActualPoints, teamPoints, ownershipPercentage)

	if err := s.ports.PortfolioWriter.UpdatePortfolioTeam(ctx, portfolioTeam); err != nil {
		log.Printf("Error updating portfolio team: %v", err)
		return err
	}
	log.Printf("Successfully updated portfolio team points")

	return nil
}
