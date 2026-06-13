module github.com/mesob-wallet/ussd

go 1.24

require (
	github.com/mesob-wallet/go-kit v0.0.0
	github.com/redis/go-redis/v9 v9.20.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.33.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace (
	github.com/mesob-wallet/events v0.0.0 => ../../shared/events
	github.com/mesob-wallet/go-kit v0.0.0 => ../../shared/go-kit
)
