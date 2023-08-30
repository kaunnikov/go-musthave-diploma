package config

import (
	"flag"
	"os"
	"strings"
)

type AppConfig struct {
	Host        string
	DatabaseDSN string
}

func LoadConfig() *AppConfig {
	cfg := &AppConfig{}
	loadFromArgs(cfg)
	loadFromENV(cfg)
	return cfg
}

func loadFromArgs(cfg *AppConfig) {
	flag.StringVar(&cfg.Host, "a", "localhost:8080", "Default Host:port")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")
	flag.Parse()
}

func loadFromENV(cfg *AppConfig) {
	envRunAddr := strings.TrimSpace(os.Getenv("SERVER_ADDRESS"))
	if envRunAddr != "" {
		cfg.Host = envRunAddr
	}

	databaseDSN := strings.TrimSpace(os.Getenv("DATABASE_DSN"))
	if databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}
}
