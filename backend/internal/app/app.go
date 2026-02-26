package app

import (
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	apppool "github.com/andrewcopp/Calcutta/backend/internal/app/pool"
	appprediction "github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
	appusermgmt "github.com/andrewcopp/Calcutta/backend/internal/app/usermanagement"

	applab "github.com/andrewcopp/Calcutta/backend/internal/app/lab"
)

type App struct {
	Analytics      *appanalytics.Service
	Lab            *applab.Service
	Bracket        *bracket.Service
	Pool           *apppool.Service
	Prediction     *appprediction.Service
	Auth           *appauth.Service
	School         *appschool.Service
	Tournament     *apptournament.Service
	UserManagement *appusermgmt.Service
}
