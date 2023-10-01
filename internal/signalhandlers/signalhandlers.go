package signalhandlers

import (
	"context"
	"net/http"
	"os"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/filestorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"go.uber.org/zap"
)

// HandleSignals listens for termination signals to gracefully shut down the application
// It saves data to a file before signaling the main routine to terminate the application
func HandleSignals(signalChan <-chan os.Signal, quitChan chan<- struct{}, storage *storage.InMemoryStorage, cfg *config.Config, sugar *zap.SugaredLogger) {
	<-signalChan
	if err := filestorage.SaveToFile(storage, cfg.FileStoragePath); err != nil {
		sugar.Errorf("Error when saving data to file: %v", err)
	}
	quitChan <- struct{}{}
}

// HandleServerErrors listens for errors from the HTTP server
// If an error is received, it logs the error and exits the application
func HandleServerErrors(errChan chan error, sugar *zap.SugaredLogger, cfg *config.Config) {
	err := <-errChan
	if err == http.ErrServerClosed {
		sugar.Info("HTTP server closed gracefully")
		return
	}
	sugar.Fatalf("Failed to start HTTP server on address %s: %s", cfg.Addr, err)
}

// HandleShutdownServer waits for a signal to shutdown the server
// It attempts to gracefully shutdown the HTTP server
func HandleShutdownServer(quitChan chan struct{}, srv *http.Server, sugar *zap.SugaredLogger) {
	<-quitChan
	if err := srv.Shutdown(context.TODO()); err != nil {
		sugar.Errorf("Server Shutdown Failed:%v", err)
	}
	sugar.Info("Server exited properly")
}
