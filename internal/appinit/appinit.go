package appinit

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/filestorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"go.uber.org/zap"
)

// InitApp initializes the application by loading the configuration and setting up the logger
// it returns a configuration object, a logger, a function to sync the logger, and possibly an error
func InitApp() (*config.Config, *zap.SugaredLogger, func(), error) {
	cfg, err := config.ParseConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error while parsing config: %w", err)
	}

	sugar, syncFunc, err := logger.InitLogger(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return cfg, sugar, syncFunc, nil
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

// InitDBStorage initializes a database storage based on the provided configuration and logger
// it returns an instance of dbstorage.StorageInterface or an error if any step in the initialization fails
func InitDBStorage(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, wg *sync.WaitGroup) (dbstorage.Interface, error) {
	dbStorage, err := dbstorage.NewDBStorage(ctx, cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	go func() {
		sugar.Info("Waiting for context to close database")
		<-ctx.Done()
		if err := dbStorage.Close(); err != nil {
			sugar.Errorf("Error while closing the database: %v", err)
		} else {
			sugar.Info("Database closed successfully")
		}
		wg.Done()
	}()

	if err := dbStorage.CreateTables(); err != nil {
		return nil, err
	}
	return dbStorage, nil
}

// InitInMemoryStorage initializes an in-memory storage based on the provided configuration and logger
// it restores any saved data and sets up automatic data saving
// it returns an instance of storage.InMemoryStorage or an error if any step in the initialization fails
func InitInMemoryStorage(cfg *config.Config, sugar *zap.SugaredLogger) (*storage.InMemoryStorage, error) {
	storage := storage.NewInMemoryStorage()
	filestorage.RestoreData(sugar, storage, cfg)
	InitDataSave(sugar, storage, cfg)
	return storage, nil
}
