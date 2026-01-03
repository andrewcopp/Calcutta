package school

import (
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	appschool "github.com/andrewcopp/Calcutta/backend/internal/app/school"
)

type Service = appschool.Service

func New(repo *dbadapters.SchoolRepository) *Service {
	return appschool.New(repo)
}
