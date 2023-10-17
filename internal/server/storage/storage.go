package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
)

// MetricStorage is an interface that provides methods for manipulating
// various types of metrics such as gauges and counters
type MetricStorage interface {
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

// InMemoryStorage is an implementation of the MetricStorage interface
// it stores the metrics in an in-memory data structure
// this implementation is thread-safe
type InMemoryStorage struct {
	updateChan chan struct{} // Channel to notify about updates
	gauges     map[string]float64
	counter    map[string]int64
	mu         sync.Mutex
}

// NewInMemoryStorage creates a new instance of InMemoryStorage and returns it
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		counter:    make(map[string]int64),
		gauges:     make(map[string]float64),
		updateChan: make(chan struct{}, 1),
	}
}

// UpdateGauge sets the current value of a gauge metric identified by its name
func (s *InMemoryStorage) UpdateGauge(name string, value float64, shouldNotify bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	s.notifyUpdate(shouldNotify)
	return nil
}

// UpdateCounter increments the value of a counter metric identified by its name
func (s *InMemoryStorage) UpdateCounter(name string, value int64, shouldNotify bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter[name] += value
	s.notifyUpdate(shouldNotify)
	return nil
}

// GetGauge fetches the current value of a gauge metric by its name from storage
func (s *InMemoryStorage) GetGauge(name string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.gauges[name]
	if !ok {
		return 0, fmt.Errorf("gauge %s not found", name)
	}

	return value, nil
}

// GetCounter fetches the current value of a counter metric by its name from storage
func (s *InMemoryStorage) GetCounter(name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter %s not found", name)
	}

	return value, nil
}

// GetMetricsData returns the stored gauges and counters metrics
func (s *InMemoryStorage) GetMetricsData() (map[string]float64, map[string]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.gauges, s.counter
}

// SetMetricsData sets the gauges and counters metrics in the storage
func (s *InMemoryStorage) SetMetricsData(gauges map[string]float64, counters map[string]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges = gauges
	s.counter = counters
}

// String provides a string representation of all the metrics in the storage
func (s *InMemoryStorage) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result strings.Builder
	result.WriteString("Counter values:\n")
	result.WriteString(utils.FormatMapSortedKeys(s.counter))
	result.WriteString("\nGauge values:\n")
	result.WriteString(utils.FormatMapSortedKeys(s.gauges))

	return result.String()
}

// GetUpdateChannel returns the update channel for this storage
// This channel is used to notify about updates in storage
func (s *InMemoryStorage) GetUpdateChannel() chan struct{} {
	return s.updateChan
}

// notifyUpdate sends a notification through updateChan if shouldNotify is true
func (s *InMemoryStorage) notifyUpdate(shouldNotify bool) {
	if shouldNotify {
		s.updateChan <- struct{}{}
	}
}
