package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
	RedisURL     string
	LedgerGRPC   string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_IQUB_HTTP_PORT", "8003"),
		GRPCPort:     getenv("MESOB_IQUB_GRPC_PORT", "9103"),
		DBURL:        getenv("MESOB_IQUB_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_iqub?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_IQUB_KAFKA_BROKERS", "localhost:9092"),
		RedisURL:     getenv("MESOB_IQUB_REDIS_URL", "localhost:6379"),
		LedgerGRPC:   getenv("MESOB_IQUB_LEDGER_GRPC", "localhost:9102"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
