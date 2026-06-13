module github.com/mesob-wallet/telegram

go 1.24

require (
	github.com/mesob-wallet/go-kit v0.0.0
	github.com/rs/zerolog v1.33.0
)

require (
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sys v0.22.0 // indirect
)

replace (
	github.com/mesob-wallet/events v0.0.0 => ../../shared/events
	github.com/mesob-wallet/go-kit v0.0.0 => ../../shared/go-kit
)
