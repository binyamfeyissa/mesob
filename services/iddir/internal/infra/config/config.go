package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
	LedgerGRPC   string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_IDDIR_HTTP_PORT", "8004"),
		GRPCPort:     getenv("MESOB_IDDIR_GRPC_PORT", "9104"),
		DBURL:        getenv("MESOB_IDDIR_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_iddir?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_IDDIR_KAFKA_BROKERS", "localhost:9092"),
		LedgerGRPC:   getenv("MESOB_IDDIR_LEDGER_GRPC", "localhost:9102"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
