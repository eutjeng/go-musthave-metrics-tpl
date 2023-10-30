package models

import (
	"context"
)

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // metric value when type is 'gauge'
	Delta *int64   `json:"delta,omitempty"` // metric value when type is 'counter'
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter that takes the value 'gauge' or 'counter'
}

type GeneralStorageInterface interface {
	// UpdateGauge updates the value of a gauge metric identified by its name.
	// If 'shouldNotify' is true, an update notification will be triggered.
	// Returns an error if the operation fails.
	UpdateGauge(ctx context.Context, name string, value float64, shouldNotify bool) error

	// UpdateCounter increments a counter metric identified by its name by a given value.
	// If 'shouldNotify' is true, an update notification will be triggered.
	// Returns an error if the operation fails.
	UpdateCounter(ctx context.Context, name string, value int64, shouldNotify bool) error

	// GetGauge retrieves the current value of a gauge metric identified by its name.
	// Returns the fetched value and an error if the operation fails.
	GetGauge(ctx context.Context, name string) (float64, error)

	// GetCounter retrieves the current value of a counter metric identified by its name.
	// Returns the fetched value and an error if the operation fails.
	GetCounter(ctx context.Context, name string) (int64, error)

	// String returns a string representation of the stored metrics.
	// This is primarily for debugging or logging purposes.
	String(ctx context.Context) string

	// SaveMetrics stores an array of metrics in the storage.
	// If 'shouldNotify' is true, an update notification will be triggered for each metric.
	// Returns an error if the operation fails.
	SaveMetrics(ctx context.Context, metrics []Metrics, shouldNotify bool) error
}
