package config

import "os"

type Config struct {
	HTTPPort    string
	Mode        string // DEMO or LIVE
	NIDEndpoint string
	MFIEndpoint string
}

func Load() Config {
	port := os.Getenv("MESOB_ADAPTER_HTTP_PORT")
	if port == "" {
		port = "8013"
	}
	return Config{
		HTTPPort:    port,
		Mode:        getEnv("MESOB_ADAPTER_MODE", "DEMO"),
		NIDEndpoint: os.Getenv("MESOB_ADAPTER_NID_ENDPOINT"),
		MFIEndpoint: os.Getenv("MESOB_ADAPTER_MFI_ENDPOINT"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
