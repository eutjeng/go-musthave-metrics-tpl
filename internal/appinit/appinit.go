package appinit

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/filestorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"go.uber.org/zap"
)

// InitApp initializes the application by loading the configuration and setting up the logger
// it returns a configuration object, a logger, and a function to sync the logger
func InitApp() (*config.Config, *zap.SugaredLogger, func()) {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Error while parsing config: %s", err)
	}

	sugar, syncFunc, err := logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}

	return cfg, sugar, syncFunc
}

// InitServer sets up and returns an HTTP server based on the provided configuration and router
func InitServer(cfg *config.Config, r http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

// InitDataSave configures the data storage mechanism based on the provided configuration
func InitDataSave(sugar *zap.SugaredLogger, storage *storage.InMemoryStorage, cfg *config.Config) {
	if cfg.StoreInterval == 0 {
		filestorage.StartSyncSave(sugar, storage, cfg.FileStoragePath)
	} else {
		filestorage.StartPeriodicSave(sugar, storage, cfg.StoreInterval, cfg.FileStoragePath)
	}
}

// InitSignalHandling sets up signal handling for graceful shutdown
// it returns channels for quit signals and system signals
func InitSignalHandling() (chan struct{}, chan os.Signal) {
	quitChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	return quitChan, signalChan
}

// StartServer launches the HTTP server in a goroutine and sends any errors to the provided channel
func StartServer(srv *http.Server, errChan chan error) {
	go func() {
		errChan <- srv.ListenAndServe()
	}()
}
