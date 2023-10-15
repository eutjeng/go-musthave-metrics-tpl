package main

import (
	"log"
	"net/http"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Error while parsing config: %s", err)
	}

	sugar, syncFunc, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}
	defer syncFunc()

	storage := storage.NewInMemoryStorage()
	r := router.SetupRouter(sugar, storage)

	err = http.ListenAndServe(cfg.Addr, r)
	if err != nil {
		sugar.Fatalf("Failed to start HTTP server on address %s: %s", cfg.Addr, err)
	}
}
