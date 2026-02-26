package bootstrap

import (
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	apppool "github.com/andrewcopp/Calcutta/backend/internal/app/pool"
	applab "github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	appprediction "github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	appusermgmt "github.com/andrewcopp/Calcutta/backend/internal/app/usermanagement"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(pool *pgxpool.Pool, cfg platform.Config, authRepo *dbadapters.AuthRepository, tm *coreauth.TokenManager) (*app.App, error) {
	dbUserRepo := dbadapters.NewUserRepository(pool)
	dbSchoolRepo := dbadapters.NewSchoolRepository(pool)
	dbTournamentRepo := dbadapters.NewTournamentRepository(pool)

	poolRepo := dbadapters.NewPoolRepository(pool)
	invitationRepo := dbadapters.NewPoolInvitationRepository(pool)
	poolService := apppool.New(apppool.Ports{
		Pools:           poolRepo,
		Portfolios:      poolRepo,
		Payouts:         poolRepo,
		OwnershipReader: poolRepo,
		ScoringRules:    poolRepo,
		TeamReader:      poolRepo,
		PoolInvitations: invitationRepo,
	})

	analyticsRepo := dbadapters.NewAnalyticsRepository(pool)
	analyticsService := appanalytics.New(analyticsRepo)

	labRepo := dbadapters.NewLabRepository(pool)
	labService := applab.New(labRepo, applab.ServiceConfig{
		DefaultNSims:      cfg.DefaultNSims,
		ExcludedEntryName: cfg.ExcludedEntryName,
	})

	predictionRepo := dbadapters.NewPredictionRepository(pool)

	a := &app.App{Bracket: appbracket.New(dbTournamentRepo)}
	a.Pool = poolService
	a.Prediction = appprediction.New(appprediction.Ports{
		Batches:    predictionRepo,
		Tournament: predictionRepo,
	})
	a.Analytics = analyticsService
	a.Lab = labService
	a.Auth = appauth.New(dbUserRepo, authRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)

	userMergeRepo := dbadapters.NewUserMergeRepository(pool)
	a.UserManagement = appusermgmt.New(appusermgmt.Ports{Merges: userMergeRepo})

	return a, nil
}
