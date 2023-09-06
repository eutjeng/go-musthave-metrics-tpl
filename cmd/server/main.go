package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/storage"
)

func main() {
	var storage storage.MetricStorage = storage.NewInMemoryStorage()
	r := chi.NewRouter()

	go func() {
		for {
			fmt.Print("\033[H\033[2J")
			fmt.Println(storage)
			time.Sleep(1 * time.Second)
		}
	}()

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.HandleUpdateMetric(storage))
	})

	err := http.ListenAndServe("localhost:8080", r)

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
