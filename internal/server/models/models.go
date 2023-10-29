package models

type Metrics struct {
	Value *float64 `json:"value,omitempty"` // metric value when type is 'gauge'
	Delta *int64   `json:"delta,omitempty"` // metric value when type is 'counter'
	ID    string   `json:"id"`              // metric name
	MType string   `json:"type"`            // parameter that takes the value 'gauge' or 'counter'
}

type GeneralStorageInterface interface {
	// UpdateGauge sets a new value for a gauge metric identified by its name
	// the function returns an error if the operation fails
	// if 'shouldNotify' is true, an update notification is triggered
	UpdateGauge(name string, value float64, shouldNotify bool) error

	// UpdateCounter increments the value of a counter metric by a given value
	// the function returns an error if the operation fails
	// if 'shouldNotify' is true, an update notification is triggered
	UpdateCounter(name string, value int64, shouldNotify bool) error

	// GetGauge fetches the current value of a gauge metric by its name
	// returns the fetched value along with an error if the operation fails
	GetGauge(name string) (float64, error)

	// GetCounter fetches the current value of a counter metric by its name
	// returns the fetched value along with an error if the operation fails
	GetCounter(name string) (int64, error)

	// String returns a stringified representation of the metrics stored
	// this is primarily useful for debugging or logging purposes
	String() string
}
