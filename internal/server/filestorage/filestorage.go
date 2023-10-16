package filestorage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"go.uber.org/zap"
)

type SerializedMetrics struct {
	Gauges  map[string]float64 `json:"gauges"`
	Counter map[string]int64   `json:"counter"`
}

// SaveToFile saves metrics to a file. Directories are created if they do not exist
func SaveToFile(cfg *config.Config, storage storage.Interface) error {
	if !cfg.Restore {
		return nil
	}

	gauges, counters := storage.GetMetricsData()

	data := SerializedMetrics{
		Gauges:  gauges,
		Counter: counters,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dir := filepath.Dir(cfg.FileStoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cfg.FileStoragePath, jsonData, 0644)
}

// LoadFromFile loads metrics from a file. If the file does not exist, no error is returned
func LoadFromFile(storage *storage.InMemoryStorage, filename string) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	data := SerializedMetrics{}
	if err = json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	storage.SetMetricsData(data.Gauges, data.Counter)
	return nil
}

func RestoreData(sugar *zap.SugaredLogger, storage *storage.InMemoryStorage, cfg *config.Config) {
	if cfg.Restore {
		if fileErr := LoadFromFile(storage, cfg.FileStoragePath); fileErr != nil {
			sugar.Errorf("Error when loading from file: %v", fileErr)
		}
	}
}

// StartSyncSave starts a goroutine that saves metrics to a file whenever an update occurs
func StartSyncSave(sugar *zap.SugaredLogger, cfg *config.Config, storage *storage.InMemoryStorage) {
	go func() {
		for range storage.GetUpdateChannel() {
			if err := SaveToFile(cfg, storage); err != nil {
				sugar.Errorf("Error when saving a file: %v", err)
			}
		}
	}()
}

// StartPeriodicSave starts a goroutine that saves metrics to a file at regular intervals
func StartPeriodicSave(sugar *zap.SugaredLogger, cfg *config.Config, storage *storage.InMemoryStorage) {
	go func() {
		ticker := time.NewTicker(cfg.StoreInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := SaveToFile(cfg, storage); err != nil {
				sugar.Errorf("Error when saving a file: %v", err)
			}
		}
	}()
}
