package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
	LedgerGRPC   string
	ScoringURL   string
	AdapterGRPC  string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_LOANS_HTTP_PORT", "8006"),
		GRPCPort:     getenv("MESOB_LOANS_GRPC_PORT", "9106"),
		DBURL:        getenv("MESOB_LOANS_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_loans?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_LOANS_KAFKA_BROKERS", "localhost:9092"),
		LedgerGRPC:   getenv("MESOB_LOANS_LEDGER_GRPC", "localhost:9102"),
		ScoringURL:   getenv("MESOB_LOANS_SCORING_URL", "http://localhost:9001"),
		AdapterGRPC:  getenv("MESOB_LOANS_ADAPTER_GRPC", "localhost:9113"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
