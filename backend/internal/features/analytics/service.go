package analytics

import (
	appanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/analytics"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service = appanalytics.Service

func New(repo ports.AnalyticsRepo) *Service {
	return appanalytics.New(repo)
}
