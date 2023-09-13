package router

import (
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(storage storage.MetricStorage) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", handlers.HandleMetricsHTML(storage))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(storage))
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.HandleGetMetric(storage))
	})

	return r
}
