package main

import (
	"database/sql"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

type Server struct {
	schoolRepo        *services.SchoolRepository
	schoolService     *services.SchoolService
	tournamentRepo    *services.TournamentRepository
	tournamentService *services.TournamentService
	calcuttaRepo      *services.CalcuttaRepository
	calcuttaService   *services.CalcuttaService
	userRepo          *services.UserRepository
	userService       *services.UserService
	bracketService    *services.BracketService
}

func NewServer(db *sql.DB) *Server {
	schoolRepo := services.NewSchoolRepository(db)
	tournamentRepo := services.NewTournamentRepository(db)
	calcuttaRepo := services.NewCalcuttaRepository(db)
	userRepo := services.NewUserRepository(db)

	schoolService := services.NewSchoolService(schoolRepo)
	tournamentService := services.NewTournamentService(tournamentRepo, schoolRepo)
	calcuttaService := services.NewCalcuttaService(services.CalcuttaServicePorts{
		CalcuttaReader:  calcuttaRepo,
		CalcuttaWriter:  calcuttaRepo,
		EntryReader:     calcuttaRepo,
		PortfolioReader: calcuttaRepo,
		PortfolioWriter: calcuttaRepo,
		RoundWriter:     calcuttaRepo,
		TeamReader:      calcuttaRepo,
	})
	userService := services.NewUserService(userRepo)
	bracketService := services.NewBracketService(tournamentRepo)

	return &Server{
		schoolRepo:        schoolRepo,
		schoolService:     schoolService,
		tournamentRepo:    tournamentRepo,
		tournamentService: tournamentService,
		calcuttaRepo:      calcuttaRepo,
		calcuttaService:   calcuttaService,
		userRepo:          userRepo,
		userService:       userService,
		bracketService:    bracketService,
	}
}
