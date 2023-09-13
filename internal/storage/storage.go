package storage

import (
	"fmt"
	"sort"
	"sync"
)

// MetricStorage is an interface that represents the methods
// that must be implemented by a metric storage
type MetricStorage interface {
	// UpdateGauge updates the value of a gauge metric
	UpdateGauge(name string, value float64) error
	// UpdateCounter increments the value of a counter metric
	UpdateCounter(name string, value int64) error
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

func (s *InMemoryStorage) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result string
	result += "Counter values:\n"
	// Sort the keys
	keysCounter := make([]string, 0, len(s.counter))
	for key := range s.counter {
		keysCounter = append(keysCounter, key)
	}
	sort.Strings(keysCounter)
	// Append sorted key-value pairs
	for _, key := range keysCounter {
		result += fmt.Sprintf("%s: %d\n", key, s.counter[key])
	}

	result += "\nGauge values:\n"
	// Sort the keys
	keysGauges := make([]string, 0, len(s.gauges))
	for key := range s.gauges {
		keysGauges = append(keysGauges, key)
	}
	sort.Strings(keysGauges)
	// Append sorted key-value pairs
	for _, key := range keysGauges {
		result += fmt.Sprintf("%s: %f\n", key, s.gauges[key])
	}

	return result
}
