package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// MemStorage - интерфейс для хранения метрик
type MemStorage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
}

// InMemoryStorage - реализация MemStorage, хранящая метрики в памяти
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

func splitPath(path string) []string {
	return strings.Split(strings.TrimPrefix(path, "/"), "/")
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func main() {
	storage := NewInMemoryStorage()

	http.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		parts := splitPath(r.URL.Path)

		if len(parts) != 4 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		metricType, metricName, metricValue := parts[1], parts[2], parts[3]

		switch metricType {
		case "gauge":
			if value, err := parseFloat(metricValue); err == nil {
				storage.UpdateGauge(metricName, value)
			} else {
				http.Error(w, "Bad request", http.StatusBadRequest)
			}
		case "counter":
			if value, err := parseInt(metricValue); err == nil {
				storage.UpdateCounter(metricName, value)
			} else {
				http.Error(w, "Bad request", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	})

	http.ListenAndServe("localhost:8080", nil)
}
