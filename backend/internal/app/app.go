package app

import (
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	applab "github.com/andrewcopp/Calcutta/backend/internal/app/lab"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
)

type App struct {
	Analytics  *appanalytics.Service
	Lab        *applab.Service
	Bracket    *bracket.Service
	Calcutta   *appcalcutta.Service
	Auth       *appauth.Service
	School     *appschool.Service
	Tournament *apptournament.Service
}
