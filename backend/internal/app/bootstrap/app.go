package bootstrap

import (
	"time"

	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/app"
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	applab "github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	appprediction "github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	coreauth "github.com/andrewcopp/Calcutta/backend/internal/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewApp(pool *pgxpool.Pool, cfg platform.Config, authRepo *dbadapters.AuthRepository, tm *coreauth.TokenManager) (*app.App, error) {
	dbUserRepo := dbadapters.NewUserRepository(pool)
	dbSchoolRepo := dbadapters.NewSchoolRepository(pool)
	dbTournamentRepo := dbadapters.NewTournamentRepository(pool)

	calcuttaRepo := dbadapters.NewCalcuttaRepository(pool)
	invitationRepo := dbadapters.NewCalcuttaInvitationRepository(pool)
	calcuttaService := appcalcutta.New(appcalcutta.Ports{
		Calcuttas:       calcuttaRepo,
		Entries:         calcuttaRepo,
		Payouts:         calcuttaRepo,
		PortfolioReader: calcuttaRepo,
		Rounds:          calcuttaRepo,
		TeamReader:      calcuttaRepo,
		Invitations:     invitationRepo,
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
	a.Calcutta = calcuttaService
	a.Prediction = appprediction.New(predictionRepo)
	a.Analytics = analyticsService
	a.Lab = labService
	a.Auth = appauth.New(dbUserRepo, authRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)

	return a, nil
}
