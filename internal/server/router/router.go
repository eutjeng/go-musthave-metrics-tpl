package router

import (
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/gzip"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbhandlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func SetupRouter(sugar *zap.SugaredLogger, storage storage.MetricStorage, dbstorage dbstorage.StorageInterface, shouldNotify bool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(gzip.WithCompression(sugar))
	r.Use(logger.WithLogging(sugar))

	r.Get("/", handlers.HandleMetricsHTML(sugar, storage))

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(sugar, storage, shouldNotify))
		r.Post("/", handlers.HandleUpdateMetric(sugar, storage, shouldNotify))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.HandleGetMetric(sugar, storage))
		r.Post("/", handlers.HandleGetMetric(sugar, storage))
	})

	r.Get("/ping", dbhandlers.PingHandler(sugar, dbstorage))

	return r
}
