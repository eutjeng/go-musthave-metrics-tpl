package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/filestorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"go.uber.org/zap"
)

func handleSignals(signalChan <-chan os.Signal, storage *storage.InMemoryStorage, cfg *config.Config, sugar *zap.SugaredLogger) {
	<-signalChan
	if err := filestorage.SaveToFile(storage, cfg.FileStoragePath); err != nil {
		sugar.Errorf("Error when saving data to file: %v", err)
	}
	os.Exit(0)
}

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
	isSyncedSaveToFile := cfg.StoreInterval == 0

	r := router.SetupRouter(sugar, storage, isSyncedSaveToFile)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if cfg.Restore {
		if fileErr := filestorage.LoadFromFile(storage, cfg.FileStoragePath); fileErr != nil {
			sugar.Errorf("Error when loading from file: %v", fileErr)
		}
	}

	if isSyncedSaveToFile {
		filestorage.StartSyncSave(sugar, storage, cfg.FileStoragePath)
	} else {
		filestorage.StartPeriodicSave(sugar, storage, cfg.StoreInterval, cfg.FileStoragePath)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go handleSignals(signalChan, storage, cfg, sugar)

	if err = srv.ListenAndServe(); err != nil {
		sugar.Fatalf("Failed to start HTTP server on address %s: %s", cfg.Addr, err)
	}

}
