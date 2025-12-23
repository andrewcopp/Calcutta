module github.com/andrewcopp/Calcutta/backend/cmd/sandbox

go 1.24

replace github.com/andrewcopp/Calcutta/backend => ../..

replace github.com/andrewcopp/Calcutta/backend/pkg/services => ../../pkg/services

replace github.com/andrewcopp/Calcutta/backend/pkg/models => ../../pkg/models

replace github.com/andrewcopp/Calcutta/backend/pkg/common => ../../pkg/common

require (
	github.com/andrewcopp/Calcutta/backend/pkg/services v0.0.0
	github.com/jackc/pgx/v5 v5.7.4
	gonum.org/v1/gonum v0.16.0
)

require (
	github.com/andrewcopp/Calcutta/backend v0.0.0 // indirect
	github.com/andrewcopp/Calcutta/backend/pkg/models v0.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
)
