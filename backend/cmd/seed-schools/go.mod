module github.com/andrewcopp/Calcutta/backend/cmd/seed-schools

go 1.24

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace github.com/andrewcopp/Calcutta/backend/pkg/common => ../../pkg/common
