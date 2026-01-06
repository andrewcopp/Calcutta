package model_catalogs

import (
	appmodelcatalogs "github.com/andrewcopp/Calcutta/backend/internal/app/model_catalogs"
)

type Service = appmodelcatalogs.Service

type ModelDescriptor = appmodelcatalogs.ModelDescriptor

func New() *Service {
	return appmodelcatalogs.New()
}
