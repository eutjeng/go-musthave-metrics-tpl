package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
)

func setupRouter(storage storage.MetricStorage) *chi.Mux {
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

func checkError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func main() {
	config.ParseFlags()
	storage := storage.NewInMemoryStorage()
	r := setupRouter(storage)

	err := http.ListenAndServe(config.FlagRunAddr, r)

	checkError(err, "Failed to start server")
}
