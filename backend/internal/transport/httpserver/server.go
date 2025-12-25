package httpserver

import (
	"database/sql"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	"github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	schoolRepo        *services.SchoolRepository
	schoolService     *services.SchoolService
	tournamentRepo    *services.TournamentRepository
	tournamentService *services.TournamentService
	calcuttaRepo      *dbadapters.CalcuttaRepository
	calcuttaService   *services.CalcuttaService
	userRepo          *services.UserRepository
	userService       *services.UserService
	app               *app.App
	analyticsRepo     *services.AnalyticsRepository
	analyticsService  *services.AnalyticsService
	authRepo          *AuthRepository
	authzRepo         *AuthorizationRepository
	tokenManager      *auth.TokenManager
	pool              *pgxpool.Pool
	cfg               platform.Config
}

func NewServer(db *sql.DB, pool *pgxpool.Pool, cfg platform.Config) *Server {
	schoolRepo := services.NewSchoolRepository(db)
	tournamentRepo := services.NewTournamentRepository(db)
	calcuttaRepo := dbadapters.NewCalcuttaRepository(pool)
	userRepo := services.NewUserRepository(db)
	analyticsRepo := services.NewAnalyticsRepository(db)
	authRepo := NewAuthRepository(pool)
	authzRepo := NewAuthorizationRepository(pool)
	dbUserRepo := dbadapters.NewUserRepository(pool)
	dbSchoolRepo := dbadapters.NewSchoolRepository(pool)
	dbTournamentRepo := dbadapters.NewTournamentRepository(pool)

	tm, _ := auth.NewTokenManager(cfg.JWTSecret, time.Duration(cfg.AccessTokenTTLSeconds)*time.Second)

	schoolService := services.NewSchoolService(schoolRepo)
	tournamentService := services.NewTournamentService(tournamentRepo, schoolRepo)
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
	userService := services.NewUserService(userRepo)
	bracketService := services.NewBracketService(dbTournamentRepo)
	a := &app.App{Bracket: appbracket.New(bracketService)}
	a.Auth = appauth.New(dbUserRepo, authRepo, authzRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)
	analyticsService := services.NewAnalyticsService(analyticsRepo)

	return &Server{
		schoolRepo:        schoolRepo,
		schoolService:     schoolService,
		tournamentRepo:    tournamentRepo,
		tournamentService: tournamentService,
		calcuttaRepo:      calcuttaRepo,
		calcuttaService:   calcuttaService,
		userRepo:          userRepo,
		userService:       userService,
		app:               a,
		analyticsRepo:     analyticsRepo,
		analyticsService:  analyticsService,
		authRepo:          authRepo,
		authzRepo:         authzRepo,
		tokenManager:      tm,
		pool:              pool,
		cfg:               cfg,
	}
}
