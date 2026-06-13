package config

import "os"

type Config struct {
	HTTPPort     string
	DBURL        string
	KafkaBrokers string
}

func Load() Config {
	port := os.Getenv("MESOB_NOTIFICATION_HTTP_PORT")
	if port == "" {
		port = "8012"
	}
	return Config{
		HTTPPort:     port,
		DBURL:        getEnv("MESOB_NOTIFICATION_DB_URL", "postgres://mesob:mesob@postgres:5432/mesob_notification"),
		KafkaBrokers: getEnv("MESOB_NOTIFICATION_KAFKA_BROKERS", "kafka:9092"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
