package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_LEDGER_HTTP_PORT", "8002"),
		GRPCPort:     getenv("MESOB_LEDGER_GRPC_PORT", "9102"),
		DBURL:        getenv("MESOB_LEDGER_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_ledger?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_LEDGER_KAFKA_BROKERS", "localhost:9092"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
