package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// MetricStorage defines an interface for storing and retrieving
// different types of metrics such as gauges and counters.
type MetricStorage interface {
	// UpdateGauge sets the current value of a gauge metric identified
	// by its name. Returns an error if the operation fails.
	UpdateGauge(name string, value float64, shouldNotify bool) error

	// UpdateCounter increments the value of a counter metric identified
	// by its name by a given value. Returns an error if the operation fails.
	UpdateCounter(name string, value int64, shouldNotify bool) error

	// GetGauge retrieves the current value of a gauge metric identified
	// by its name. Returns the value and an error if the operation fails.
	GetGauge(name string) (float64, error)

	// GetCounter retrieves the current value of a counter metric identified
	// by its name. Returns the value and an error if the operation fails.
	GetCounter(name string) (int64, error)

	// String returns a string representation of the stored metrics,
	// useful for debugging or logging.
	String() string
}

// InMemoryStorage is an implementation of the MetricStorage interface
// that stores the metrics in memory
type InMemoryStorage struct {
	updateChan chan struct{}
	gauges     map[string]float64
	counter    map[string]int64
	mu         sync.RWMutex
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
// it locks the storage before updating and unlocks it afterward
func (s *InMemoryStorage) UpdateGauge(name string, value float64, shouldNotify bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	s.notifyUpdate(shouldNotify)
	return nil
}

// UpdateCounter increments the value of a counter metric identified by its name
// it locks the storage before updating and unlocks it afterward
func (s *InMemoryStorage) UpdateCounter(name string, value int64, shouldNotify bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter[name] += value
	s.notifyUpdate(shouldNotify)
	return nil
}

// GetGauge fetches the current value of a gauge metric by its name from storage
// it acquires a read lock before fetching and releases it afterward
func (s *InMemoryStorage) GetGauge(name string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.gauges[name]
	if !ok {
		return 0, fmt.Errorf("gauge %s not found", name)
	}

	return value, nil
}

// GetCounter fetches the current value of a counter metric by its name from storage
// it acquires a read lock before fetching and releases it afterward
func (s *InMemoryStorage) GetCounter(name string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter %s not found", name)
	}

	return value, nil
}

func (s *InMemoryStorage) GetMetricsData() (map[string]float64, map[string]int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.gauges, s.counter
}

func (s *InMemoryStorage) SetMetricsData(gauges map[string]float64, counters map[string]int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges = gauges
	s.counter = counters
}

func (s *InMemoryStorage) notifyUpdate(shouldNotify bool) {
	if shouldNotify {
		s.updateChan <- struct{}{}
	}
}

func (s *InMemoryStorage) GetUpdateChannel() chan struct{} {
	return s.updateChan
}

// getSortedKeys takes a map and returns its keys sorted as a slice of strings
func getSortedKeys(m interface{}) []string {
	var keys []string
	v := reflect.ValueOf(m)

	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			keys = append(keys, key.String())
		}

		sort.Strings(keys)
	}

	return keys
}

// formatMapSortedKeys formats the key-value pairs of a map to a string,
// with keys sorted
func (s *InMemoryStorage) formatMapSortedKeys(m interface{}) string {
	var result strings.Builder
	keys := getSortedKeys(m)

	for _, key := range keys {
		switch mapType := m.(type) {
		case map[string]int64:
			result.WriteString(fmt.Sprintf("%s: %d\n", key, mapType[key]))
		case map[string]float64:
			result.WriteString(fmt.Sprintf("%s: %f\n", key, mapType[key]))
		}
	}
	return result.String()
}

// String provides a string representation of all the metrics in the storage
// it locks the storage before generating the string and unlocks it afterward
func (s *InMemoryStorage) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result strings.Builder
	result.WriteString("Counter values:\n")
	result.WriteString(s.formatMapSortedKeys(s.counter))
	result.WriteString("\nGauge values:\n")
	result.WriteString(s.formatMapSortedKeys(s.gauges))

	return result.String()
}
