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

	sugar, syncFunc, err := logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}
	defer syncFunc()

	storage := storage.NewInMemoryStorage()
	r := router.SetupRouter(sugar, storage)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err = srv.ListenAndServe(); err != nil {
		sugar.Fatalf("Failed to start HTTP server on address %s: %s", cfg.Addr, err)
	}
}
