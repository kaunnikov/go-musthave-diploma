package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"kaunnikov/internal/auth"
	"kaunnikov/internal/compression"
	"kaunnikov/internal/config"
)

type app struct {
	*chi.Mux
	cfg *config.AppConfig
}

func NewApp(cfg *config.AppConfig) *app {
	a := &app{
		chi.NewRouter(),
		cfg,
	}
	a.registerRoutes()
	return a
}

func (m *app) registerRoutes() {
	m.Use(middleware.RequestID)
	m.Use(middleware.RealIP)
	m.Use(middleware.Logger)
	m.Use(middleware.Recoverer)
	m.Use(compression.CustomCompression)

	m.Post("/api/user/register", m.RegistrationHandler)
	m.Post("/api/user/login", m.LoginHandler)

	// Группа роутов с проверкой авторизации
	m.With(auth.CustomAuthMiddleware).Group(func(r chi.Router) {
		r.Post("/api/user/orders", m.NewOrderHandler)
		r.Get("/api/user/orders", m.GetOrdersHandler)

		r.Get("/api/user/balance", m.UserBalanceHandler)
		r.Post("/api/user/balance/withdraw", m.WithdrawHandler)
		r.Get("/api/user/withdrawals", m.GetWithdrawsHandler)
	})
}
