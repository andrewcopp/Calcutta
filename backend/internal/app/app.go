package app

import (
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
)

type App struct {
	Bracket    *bracket.Service
	Auth       *appauth.Service
	School     *appschool.Service
	Tournament *apptournament.Service
}
