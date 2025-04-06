module github.com/andrewcopp/Calcutta/backend/pkg/services

go 1.24

require (
	github.com/andrewcopp/Calcutta/backend/pkg/models v0.0.0
	github.com/google/uuid v1.6.0
)

replace github.com/andrewcopp/Calcutta/backend/pkg/common => ../common

replace github.com/andrewcopp/Calcutta/backend/pkg/models => ../models
