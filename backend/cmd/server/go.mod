module github.com/andrewcopp/Calcutta/backend/cmd/server

go 1.24

require (
	github.com/andrewcopp/Calcutta/backend/pkg/models v0.0.0
	github.com/andrewcopp/Calcutta/backend/pkg/services v0.0.0
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
)

require github.com/google/uuid v1.6.0 // indirect

replace github.com/andrewcopp/Calcutta/backend/pkg/services => ../../pkg/services

replace github.com/andrewcopp/Calcutta/backend/pkg/models => ../../pkg/models
