package router

import (
	"context"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/gzip"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbhandlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func SetupRouter(ctx context.Context, sugar *zap.SugaredLogger, store models.GeneralStorageInterface, shouldNotify bool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(gzip.WithCompression(sugar))
	r.Use(logger.WithLogging(sugar))

	r.Get("/", handlers.HandleMetricsHTML(ctx, sugar, store))

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(ctx, sugar, store, shouldNotify))
		r.Post("/", handlers.HandleUpdateMetric(ctx, sugar, store, shouldNotify))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.HandleGetMetric(ctx, sugar, store))
		r.Post("/", handlers.HandleGetMetric(ctx, sugar, store))
	})

	if s, ok := store.(dbstorage.Interface); ok {
		r.Get("/ping", dbhandlers.PingHandler(sugar, s))
	} else {
		sugar.Warn("Store does not support dbhandlers")
	}

	return r
}
