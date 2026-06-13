package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort      string
	BotToken      string
	WebhookSecret string
	GatewayURL    string
}

func Load() Config {
	port := os.Getenv("MESOB_TELEGRAM_HTTP_PORT")
	if port == "" {
		port = "8011"
	}
	_ = strconv.Itoa(0) // keep import
	return Config{
		HTTPPort:      port,
		BotToken:      os.Getenv("MESOB_TELEGRAM_BOT_TOKEN"),
		WebhookSecret: os.Getenv("MESOB_TELEGRAM_WEBHOOK_SECRET"),
		GatewayURL:    getEnv("MESOB_TELEGRAM_GATEWAY_URL", "http://gateway:8000"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
