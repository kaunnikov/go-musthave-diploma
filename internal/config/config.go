package config

import (
	"flag"
	"os"
	"strings"
)

type AppConfig struct {
	Host                 string
	DatabaseDSN          string
	AccrualSystemAddress string
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
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "ACCRUAL SYSTEM ADDRESS")

	flag.Parse()
}

func loadFromENV(cfg *AppConfig) {
	envRunAddr := strings.TrimSpace(os.Getenv("RUN_ADDRESS"))
	if envRunAddr != "" {
		cfg.Host = envRunAddr
	}

	databaseDSN := strings.TrimSpace(os.Getenv("DATABASE_URI"))
	if databaseDSN != "" {
		cfg.DatabaseDSN = databaseDSN
	}

	accrualSystemAddress := strings.TrimSpace(os.Getenv("ACCRUAL_SYSTEM_ADDRESS"))
	if accrualSystemAddress != "" {
		cfg.AccrualSystemAddress = accrualSystemAddress
	}
}
