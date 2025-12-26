package bootstrap

import (
	"database/sql"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(db *sql.DB, pool *pgxpool.Pool, cfg platform.Config, authRepo *dbadapters.AuthRepository, authzRepo *dbadapters.AuthorizationRepository) (*app.App, *coreauth.TokenManager, error) {
	dbUserRepo := dbadapters.NewUserRepository(pool)
	dbSchoolRepo := dbadapters.NewSchoolRepository(pool)
	dbTournamentRepo := dbadapters.NewTournamentRepository(pool)

	tm, err := coreauth.NewTokenManager(cfg.JWTSecret, time.Duration(cfg.AccessTokenTTLSeconds)*time.Second)
	if err != nil {
		return nil, nil, err
	}

	calcuttaRepo := dbadapters.NewCalcuttaRepository(pool)
	calcuttaService := services.NewCalcuttaService(services.CalcuttaServicePorts{
		CalcuttaReader:  calcuttaRepo,
		CalcuttaWriter:  calcuttaRepo,
		EntryReader:     calcuttaRepo,
		EntryWriter:     calcuttaRepo,
		PayoutReader:    calcuttaRepo,
		PortfolioReader: calcuttaRepo,
		PortfolioWriter: calcuttaRepo,
		RoundWriter:     calcuttaRepo,
		TeamReader:      calcuttaRepo,
	})

	analyticsRepo := services.NewAnalyticsRepository(db)
	analyticsService := services.NewAnalyticsService(analyticsRepo)

	bracketService := services.NewBracketService(dbTournamentRepo)

	a := &app.App{Bracket: appbracket.New(bracketService)}
	a.Calcutta = appcalcutta.New(calcuttaService)
	a.Analytics = appanalytics.New(analyticsService)
	a.Auth = appauth.New(dbUserRepo, authRepo, authzRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)

	return a, tm, nil
}
