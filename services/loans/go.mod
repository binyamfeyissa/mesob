module github.com/mesob-wallet/loans

go 1.24

require (
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v5 v5.6.0
	github.com/mesob-wallet/go-kit v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)

replace (
	github.com/mesob-wallet/events v0.0.0 => ../../shared/events
	github.com/mesob-wallet/go-kit v0.0.0 => ../../shared/go-kit
)
