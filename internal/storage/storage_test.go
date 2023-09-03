package storage

import (
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	tests := []struct {
		name           string
		counterUpdates map[string]int64
		gaugeUpdates   map[string]float64
	}{
		{
			name: "Test1",
			counterUpdates: map[string]int64{
				"counter1": 1,
				"counter2": 2,
			},
			gaugeUpdates: map[string]float64{
				"gauge1": 1.1,
				"gauge2": 2.2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryStorage()

			// Update counters
			for name, value := range tt.counterUpdates {
				err := s.UpdateCounter(name, value)
				if err != nil {
					t.Errorf("UpdateCounter() error = %v", err)
				}
			}

			// Update gauges
			for name, value := range tt.gaugeUpdates {
				err := s.UpdateGauge(name, value)
				if err != nil {
					t.Errorf("UpdateGauge() error = %v", err)
				}
			}

			// Check counters
			for name, value := range tt.counterUpdates {
				if s.counter[name] != value {
					t.Errorf("Expected counter %s to be %d, got %d", name, value, s.counter[name])
				}
			}

			// Check gauges
			for name, value := range tt.gaugeUpdates {
				if s.gauges[name] != value {
					t.Errorf("Expected gauge %s to be %f, got %f", name, value, s.gauges[name])
				}
			}
		})
	}
}
