package main

import (
	"kaunnikov/internal/accrual"
	"kaunnikov/internal/app"
	"kaunnikov/internal/config"
	"kaunnikov/internal/db"
	"kaunnikov/internal/logging"
	"log"
	"net/http"
	"time"
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

	// Раз в 5 секунд забираем из БД необработанные заказы и обрабатываем их
	go func() {
		for {
			accrual.CheckOrders(cfg.AccrualSystemAddress + "/api/orders/")
			time.Sleep(time.Second * 5)
		}
	}()

	logging.Infof("Running server on %s", cfg.Host)
	logging.Fatalf("cannot listen and serve: %s", http.ListenAndServe(cfg.Host, newApp))

}
