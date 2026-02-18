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
	invitationRepo := dbadapters.NewCalcuttaInvitationRepository(pool)
	calcuttaService := appcalcutta.New(appcalcutta.Ports{
		CalcuttaReader:   calcuttaRepo,
		CalcuttaWriter:   calcuttaRepo,
		EntryReader:      calcuttaRepo,
		EntryWriter:      calcuttaRepo,
		PayoutReader:     calcuttaRepo,
		PayoutWriter:     calcuttaRepo,
		PortfolioReader:  calcuttaRepo,
		RoundReader:      calcuttaRepo,
		RoundWriter:      calcuttaRepo,
		TeamReader:       calcuttaRepo,
		InvitationReader: invitationRepo,
		InvitationWriter: invitationRepo,
	})

	analyticsRepo := dbadapters.NewAnalyticsRepository(pool)
	analyticsService := appanalytics.New(analyticsRepo)

	labRepo := dbadapters.NewLabRepository(pool)
	labService := applab.NewWithPipelineRepo(labRepo, applab.ServiceConfig{
		DefaultNSims:      cfg.DefaultNSims,
		ExcludedEntryName: cfg.ExcludedEntryName,
	})

	a := &app.App{Bracket: appbracket.New(dbTournamentRepo)}
	a.Calcutta = calcuttaService
	a.Analytics = analyticsService
	a.Lab = labService
	a.Auth = appauth.New(dbUserRepo, authRepo, authzRepo, tm, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	a.School = appschool.New(dbSchoolRepo)
	a.Tournament = apptournament.New(dbTournamentRepo)

	return a, tm, nil
}
