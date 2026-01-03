package ml_analytics

import (
	appmlanalytics "github.com/andrewcopp/Calcutta/backend/internal/app/ml_analytics"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
)

type Service = appmlanalytics.Service

func New(repo ports.MLAnalyticsRepository) *Service {
	return appmlanalytics.New(repo)
}
