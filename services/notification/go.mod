module github.com/mesob-wallet/notification

go 1.25.0

require (
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/jackc/pgx/v5 v5.10.0
	github.com/mesob-wallet/go-kit v0.0.0
	github.com/rs/zerolog v1.33.0
	github.com/segmentio/kafka-go v0.4.51
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace (
	github.com/mesob-wallet/events v0.0.0 => ../../shared/events
	github.com/mesob-wallet/go-kit v0.0.0 => ../../shared/go-kit
)
