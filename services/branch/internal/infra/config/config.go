package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
	RedisURL     string
	LedgerGRPC   string
	IdentityGRPC string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_BRANCH_HTTP_PORT", "8008"),
		GRPCPort:     getenv("MESOB_BRANCH_GRPC_PORT", "9108"),
		DBURL:        getenv("MESOB_BRANCH_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_branch?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_BRANCH_KAFKA_BROKERS", "localhost:9092"),
		RedisURL:     getenv("MESOB_BRANCH_REDIS_URL", "localhost:6379"),
		LedgerGRPC:   getenv("MESOB_BRANCH_LEDGER_GRPC", "localhost:9102"),
		IdentityGRPC: getenv("MESOB_BRANCH_IDENTITY_GRPC", "localhost:9101"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
