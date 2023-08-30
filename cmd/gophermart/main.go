package main

import (
	"kaunnikov/internal/app"
	"kaunnikov/internal/config"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
	"log"
	"net/http"
)

func main() {

	cfg := config.LoadConfig()

	if err := logging.Init(); err != nil {
		log.Fatalf("logger don't Run!: %s", err)
	}

	if err := db.Init(cfg); err != nil {
		logging.Fatalf("db don't init: %s", err)
	}

	newApp := app.NewApp(cfg)

	logging.Infof("Running server on %s", cfg.Host)
	logging.Fatalf("cannot listen and serve: %s", http.ListenAndServe(cfg.Host, newApp))
}
