package bootstrap

import (
	"context"
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/app/algorithm_registry"
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/app/lab_candidates"
	"github.com/andrewcopp/Calcutta/backend/internal/app/ml_analytics"
	"github.com/andrewcopp/Calcutta/backend/internal/app/model_catalogs"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(pool *pgxpool.Pool, cfg platform.Config, authRepo *dbadapters.AuthRepository, authzRepo *dbadapters.AuthorizationRepository) (*app.App, *coreauth.TokenManager, error) {
	dbUserRepo := dbadapters.NewUserRepository(pool)
	dbSchoolRepo := dbadapters.NewSchoolRepository(pool)
	dbTournamentRepo := dbadapters.NewTournamentRepository(pool)

	var tm *coreauth.TokenManager
	if cfg.AuthMode != "cognito" {
		created, err := coreauth.NewTokenManager(cfg.JWTSecret, time.Duration(cfg.AccessTokenTTLSeconds)*time.Second)
		if err != nil {
			return nil, nil, err
		}
		tm = created
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
	if err := algorithm_registry.SyncToDatabase(context.Background(), pool, algorithm_registry.RegisteredAlgorithms()); err != nil {
		return nil, nil, err
	}

	mlAnalyticsRepo := dbadapters.NewMLAnalyticsRepository(pool)
	mlAnalyticsService := ml_analytics.New(mlAnalyticsRepo)

	labCandidatesRepo := dbadapters.NewLabCandidatesRepository(pool)
	labCandidatesService := lab_candidates.New(labCandidatesRepo)

	a := &app.App{Bracket: appbracket.New(dbTournamentRepo)}
	a.Calcutta = calcuttaService
	a.Analytics = analyticsService
	a.LabCandidates = labCandidatesService
	a.MLAnalytics = mlAnalyticsService
	a.ModelCatalogs = model_catalogs.New()
	a.Auth = appauth.New(dbUserRepo, authRepo, authzRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)

	return a, tm, nil
}
