package bootstrap

import (
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/features/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/features/auth"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/features/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/features/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/features/ml_analytics"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/features/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/features/tournament"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(pool *pgxpool.Pool, cfg platform.Config, authRepo *dbadapters.AuthRepository, authzRepo *dbadapters.AuthorizationRepository) (*app.App, *coreauth.TokenManager, error) {
	dbUserRepo := dbadapters.NewUserRepository(pool)
	dbSchoolRepo := dbadapters.NewSchoolRepository(pool)
	dbTournamentRepo := dbadapters.NewTournamentRepository(pool)

	tm, err := coreauth.NewTokenManager(cfg.JWTSecret, time.Duration(cfg.AccessTokenTTLSeconds)*time.Second)
	if err != nil {
		return nil, nil, err
	}

	calcuttaRepo := dbadapters.NewCalcuttaRepository(pool)
	calcuttaService := appcalcutta.New(appcalcutta.Ports{
		CalcuttaReader:  calcuttaRepo,
		CalcuttaWriter:  calcuttaRepo,
		EntryReader:     calcuttaRepo,
		EntryWriter:     calcuttaRepo,
		PayoutReader:    calcuttaRepo,
		PortfolioReader: calcuttaRepo,
		RoundReader:     calcuttaRepo,
		RoundWriter:     calcuttaRepo,
		TeamReader:      calcuttaRepo,
	})

	analyticsRepo := dbadapters.NewAnalyticsRepository(pool)
	analyticsService := appanalytics.New(analyticsRepo)

	mlAnalyticsRepo := dbadapters.NewMLAnalyticsRepository(pool)
	mlAnalyticsService := ml_analytics.New(mlAnalyticsRepo)

	a := &app.App{Bracket: appbracket.New(dbTournamentRepo)}
	a.Calcutta = calcuttaService
	a.Analytics = analyticsService
	a.MLAnalytics = mlAnalyticsService
	a.Auth = appauth.New(dbUserRepo, authRepo, authzRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)

	return a, tm, nil
}
