package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
	LedgerGRPC   string
	LedgerURL    string
	FraudURL     string
	IdentityURL  string
	AdapterGRPC  string
	IdentityGRPC string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_PAYMENTS_HTTP_PORT", "8007"),
		GRPCPort:     getenv("MESOB_PAYMENTS_GRPC_PORT", "9107"),
		DBURL:        getenv("MESOB_PAYMENTS_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_payments?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_PAYMENTS_KAFKA_BROKERS", "localhost:9092"),
		LedgerGRPC:   getenv("MESOB_PAYMENTS_LEDGER_GRPC", "localhost:9102"),
		LedgerURL:    getenv("MESOB_PAYMENTS_LEDGER_URL", "http://localhost:8002"),
		FraudURL:     getenv("MESOB_PAYMENTS_FRAUD_URL", "http://localhost:8009"),
		IdentityURL:  getenv("MESOB_PAYMENTS_IDENTITY_URL", "http://localhost:8001"),
		AdapterGRPC:  getenv("MESOB_PAYMENTS_ADAPTER_GRPC", "localhost:9113"),
		IdentityGRPC: getenv("MESOB_PAYMENTS_IDENTITY_GRPC", "localhost:9101"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
