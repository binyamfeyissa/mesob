package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	RedisURL     string
	KafkaBrokers string
	LedgerGRPC   string
	AdapterGRPC  string
	NotifyGRPC   string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_IDENTITY_HTTP_PORT", "8001"),
		GRPCPort:     getenv("MESOB_IDENTITY_GRPC_PORT", "9101"),
		DBURL:        getenv("MESOB_IDENTITY_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_identity?sslmode=disable"),
		RedisURL:     getenv("MESOB_IDENTITY_REDIS_URL", "localhost:6379"),
		KafkaBrokers: getenv("MESOB_IDENTITY_KAFKA_BROKERS", "localhost:9092"),
		LedgerGRPC:   getenv("MESOB_IDENTITY_LEDGER_GRPC", "localhost:9102"),
		AdapterGRPC:  getenv("MESOB_IDENTITY_ADAPTER_GRPC", "localhost:9113"),
		NotifyGRPC:   getenv("MESOB_IDENTITY_NOTIFY_GRPC", "localhost:9112"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
