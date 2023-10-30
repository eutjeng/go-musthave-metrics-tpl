package router

import (
	"context"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/gzip"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/hash"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/retry"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbhandlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func SetupRouter(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, store models.GeneralStorageInterface, shouldNotify bool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(gzip.WithCompression(sugar))
	r.Use(logger.WithLogging(sugar))
	r.Use(hash.WithHashing(cfg, sugar))
	r.Use(retry.WithRetry(ctx, cfg, sugar))

	r.Get("/", handlers.HandleMetricsHTML(ctx, cfg, sugar, store))

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(ctx, cfg, sugar, store, shouldNotify))
		r.Post("/", handlers.HandleUpdateMetric(ctx, cfg, sugar, store, shouldNotify))
	})

	r.Route("/updates", func(r chi.Router) {
		r.Post("/", handlers.HandleSaveMetrics(ctx, cfg, sugar, store, shouldNotify))
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.HandleGetMetric(ctx, cfg, sugar, store))
		r.Post("/", handlers.HandleGetMetric(ctx, cfg, sugar, store))
	})

	if s, ok := store.(dbstorage.Interface); ok {
		r.Get("/ping", dbhandlers.PingHandler(sugar, s))
	} else {
		sugar.Warn("Store does not support /ping route")
	}

	return r
}
