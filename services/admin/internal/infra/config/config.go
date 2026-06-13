package config

import "os"

type Config struct {
	HTTPPort     string
	GRPCPort     string
	DBURL        string
	KafkaBrokers string
	RedisURL     string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MESOB_ADMIN_HTTP_PORT", "8009"),
		GRPCPort:     getenv("MESOB_ADMIN_GRPC_PORT", "9109"),
		DBURL:        getenv("MESOB_ADMIN_DB_URL", "postgres://mesob:mesob@localhost:5432/mesob_admin?sslmode=disable"),
		KafkaBrokers: getenv("MESOB_ADMIN_KAFKA_BROKERS", "localhost:9092"),
		RedisURL:     getenv("MESOB_ADMIN_REDIS_URL", "localhost:6379"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
