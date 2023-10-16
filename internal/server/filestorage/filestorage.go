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
func SaveToFile(storage storage.Interface, filename string) error {
	gauges, counters := storage.GetMetricsData()

	data := SerializedMetrics{
		Gauges:  gauges,
		Counter: counters,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
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
func StartSyncSave(sugar *zap.SugaredLogger, storage *storage.InMemoryStorage, filename string) {
	go func() {
		for range storage.GetUpdateChannel() {
			if err := SaveToFile(storage, filename); err != nil {
				sugar.Errorf("Error when saving a file: %v", err)
			}
		}
	}()
}

// StartPeriodicSave starts a goroutine that saves metrics to a file at regular intervals
func StartPeriodicSave(sugar *zap.SugaredLogger, storage *storage.InMemoryStorage, interval time.Duration, filename string) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := SaveToFile(storage, filename); err != nil {
				sugar.Errorf("Error when saving a file: %v", err)
			}
		}
	}()
}
