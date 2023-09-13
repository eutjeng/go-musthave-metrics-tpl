package main

import (
	"log"
	"net/http"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
)

func main() {
	cfg, err := config.ParseConfig()

	if err != nil {
		log.Fatalf("Error while parsing config: %s", err)
	}

	storage := storage.NewInMemoryStorage()
	r := router.SetupRouter(storage)

	err = http.ListenAndServe(cfg.Addr, r)

	if err != nil {
		log.Fatalf("Failed to start HTTP server on address %s: %s", cfg.Addr, err)
	}
}
