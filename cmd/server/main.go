package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
)

func main() {
	config.ParseFlags()
	var storage storage.MetricStorage = storage.NewInMemoryStorage()
	r := chi.NewRouter()

	// HTML route for metrics
	r.Get("/", handlers.HandleMetricsHTML(storage))

	// Group metric update routes
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(storage))
	})

	// Group value retrieval routes
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", handlers.HandleGetMetric(storage))
	})

	err := http.ListenAndServe(config.FlagRunAddr, r)

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
