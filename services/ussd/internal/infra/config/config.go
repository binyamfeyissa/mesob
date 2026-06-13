package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	RedisURL     string
	IdentityGRPC string
	PaymentsGRPC string
	LoansGRPC    string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_USSD_HTTP_PORT", "8010"),
		GRPCPort:     getenv("MESOB_USSD_GRPC_PORT", "9110"),
		RedisURL:     getenv("MESOB_USSD_REDIS_URL", "localhost:6379"),
		IdentityGRPC: getenv("MESOB_USSD_IDENTITY_GRPC", "localhost:9101"),
		PaymentsGRPC: getenv("MESOB_USSD_PAYMENTS_GRPC", "localhost:9107"),
		LoansGRPC:    getenv("MESOB_USSD_LOANS_GRPC", "localhost:9106"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
