package storage

import (
	"fmt"
	"sort"
	"sync"
)

// MetricStorage defines an interface for storing and retrieving
// different types of metrics such as gauges and counters.
type MetricStorage interface {
	// UpdateGauge sets the current value of a gauge metric identified
	// by its name. Returns an error if the operation fails.
	UpdateGauge(name string, value float64) error

	// UpdateCounter increments the value of a counter metric identified
	// by its name by a given value. Returns an error if the operation fails.
	UpdateCounter(name string, value int64) error

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
	mu      sync.Mutex
	counter map[string]int64
	gauges  map[string]float64
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		counter: make(map[string]int64),
		gauges:  make(map[string]float64),
	}
}

func (s *InMemoryStorage) UpdateGauge(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	return nil
}

func (s *InMemoryStorage) UpdateCounter(name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter[name] += value
	return nil
}

func (s *InMemoryStorage) GetGauge(name string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.gauges[name]
	if !ok {
		return 0, fmt.Errorf("gauge %s not found", name)
	}

	return value, nil
}

func (s *InMemoryStorage) GetCounter(name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.counter[name]
	if !ok {
		return 0, fmt.Errorf("counter %s not found", name)
	}

	return value, nil
}

func (s *InMemoryStorage) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result string
	result += "Counter values:\n"

	keysCounter := make([]string, 0, len(s.counter))
	for key := range s.counter {
		keysCounter = append(keysCounter, key)
	}

	sort.Strings(keysCounter)
	for _, key := range keysCounter {
		result += fmt.Sprintf("%s: %d\n", key, s.counter[key])
	}

	result += "\nGauge values:\n"
	keysGauges := make([]string, 0, len(s.gauges))
	for key := range s.gauges {
		keysGauges = append(keysGauges, key)
	}

	sort.Strings(keysGauges)
	for _, key := range keysGauges {
		result += fmt.Sprintf("%s: %f\n", key, s.gauges[key])
	}

	return result
}
