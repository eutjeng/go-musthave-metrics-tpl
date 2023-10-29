package router

import (
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/gzip"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbhandlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func SetupRouter(sugar *zap.SugaredLogger, store models.GeneralStorageInterface, shouldNotify bool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(gzip.WithCompression(sugar))
	r.Use(logger.WithLogging(sugar))

	r.Get("/", handlers.HandleMetricsHTML(sugar, store))

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(sugar, store, shouldNotify))
		r.Post("/", handlers.HandleUpdateMetric(sugar, store, shouldNotify))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.HandleGetMetric(sugar, store))
		r.Post("/", handlers.HandleGetMetric(sugar, store))
	})

	if s, ok := store.(dbstorage.Interface); ok {
		sugar.Info("Setting up route for /ping")
		r.Get("/ping", dbhandlers.PingHandler(sugar, s))
	} else {
		sugar.Warn("Store does not support ping")
	}

	return r
}
