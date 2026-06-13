package config

import (
	"os"
)

type Config struct {
	HTTPPort        string
	IdentityURL     string
	LedgerURL       string
	PaymentsURL     string
	LoansURL        string
	IqubURL         string
	IddirURL        string
	AgentURL        string
	BranchURL       string
	AdminURL        string
	UssdURL         string
	NotificationURL string
	RedisURL        string
	CORSOrigin      string
}

func Load() Config {
	return Config{
		HTTPPort:        getenv("MESOB_GATEWAY_HTTP_PORT", "8000"),
		IdentityURL:     getenv("MESOB_GATEWAY_IDENTITY_URL", "http://localhost:8001"),
		LedgerURL:       getenv("MESOB_GATEWAY_LEDGER_URL", "http://localhost:8002"),
		PaymentsURL:     getenv("MESOB_GATEWAY_PAYMENTS_URL", "http://localhost:8007"),
		LoansURL:        getenv("MESOB_GATEWAY_LOANS_URL", "http://localhost:8006"),
		IqubURL:         getenv("MESOB_GATEWAY_IQUB_URL", "http://localhost:8003"),
		IddirURL:        getenv("MESOB_GATEWAY_IDDIR_URL", "http://localhost:8004"),
		AgentURL:        getenv("MESOB_GATEWAY_AGENT_URL", "http://localhost:8005"),
		BranchURL:       getenv("MESOB_GATEWAY_BRANCH_URL", "http://localhost:8008"),
		AdminURL:        getenv("MESOB_GATEWAY_ADMIN_URL", "http://localhost:8009"),
		UssdURL:         getenv("MESOB_GATEWAY_USSD_URL", "http://localhost:8010"),
		NotificationURL: getenv("MESOB_GATEWAY_NOTIFICATION_URL", "http://localhost:8012"),
		RedisURL:        getenv("MESOB_GATEWAY_REDIS_URL", "localhost:6379"),
		CORSOrigin:      getenv("MESOB_GATEWAY_CORS_ORIGIN", "http://localhost:3000"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
