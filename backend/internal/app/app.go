package app

import (
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	appauth "github.com/andrewcopp/Calcutta/backend/internal/app/auth"
	"github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	appcalcutta "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/app/lab_candidates"
	"github.com/andrewcopp/Calcutta/backend/internal/app/ml_analytics"
	"github.com/andrewcopp/Calcutta/backend/internal/app/model_catalogs"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
	"github.com/andrewcopp/Calcutta/backend/internal/app/strategy_runs"
	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_evaluations"
	"github.com/andrewcopp/Calcutta/backend/internal/app/suite_scenarios"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
)

type App struct {
	Analytics        *appanalytics.Service
	LabCandidates    *lab_candidates.Service
	MLAnalytics      *ml_analytics.Service
	ModelCatalogs    *model_catalogs.Service
	Bracket          *bracket.Service
	Calcutta         *appcalcutta.Service
	SuiteScenarios   *suite_scenarios.Service
	SuiteEvaluations *suite_evaluations.Service
	StrategyRuns     *strategy_runs.Service
	Auth             *appauth.Service
	School           *appschool.Service
	Tournament       *apptournament.Service
}
