package app

import (
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/features/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/features/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/features/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/features/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/features/ml_analytics"
	"github.com/andrewcopp/Calcutta/backend/internal/features/model_catalogs"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/features/school"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/features/tournament"
)

type App struct {
	Analytics     *appanalytics.Service
	MLAnalytics   *ml_analytics.Service
	ModelCatalogs *model_catalogs.Service
	Bracket       *bracket.Service
	Calcutta      *appcalcutta.Service
	Auth          *appauth.Service
	School        *appschool.Service
	Tournament    *apptournament.Service
}
