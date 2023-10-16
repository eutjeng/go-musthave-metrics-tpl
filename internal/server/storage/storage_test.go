package storage

import (
	"context"
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	tests := []struct {
		gaugeUpdates   map[string]float64
		counterUpdates map[string]int64
		name           string
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

			t.Run("UpdateAndCheckCounters", func(t *testing.T) {
				for name, value := range tt.counterUpdates {
					err := s.UpdateCounter(context.TODO(), name, value, false)
					if err != nil {
						t.Errorf("UpdateCounter() error = %v", err)
					}
				}

				for name, value := range tt.counterUpdates {
					retrievedValue, err := s.GetCounter(context.TODO(), name)
					if err != nil {
						t.Errorf("GetCounter() error = %v", err)
					}

					if retrievedValue != value {
						t.Errorf("Expected counter %s to be %d, got %d", name, value, retrievedValue)
					}
				}
			})

			t.Run("UpdateAndCheckGauges", func(t *testing.T) {
				for name, value := range tt.gaugeUpdates {
					err := s.UpdateGauge(context.TODO(), name, value, false)
					if err != nil {
						t.Errorf("UpdateGauge() error = %v", err)
					}
				}

				for name, value := range tt.gaugeUpdates {
					retrievedValue, err := s.GetGauge(context.TODO(), name)
					if err != nil {
						t.Errorf("GetGauge() error = %v", err)
					}

					if retrievedValue != value {
						t.Errorf("Expected gauge %s to be %f, got %f", name, value, retrievedValue)
					}
				}
			})

			t.Run("TestNonExistentCounter", func(t *testing.T) {
				_, err := s.GetCounter(context.TODO(), "nonExistentCounter")
				if err == nil {
					t.Errorf("Expected an error for non-existent counter")
				}
			})

			t.Run("TestNonExistentGauge", func(t *testing.T) {
				_, err := s.GetGauge(context.TODO(), "nonExistentGauge")
				if err == nil {
					t.Errorf("Expected an error for non-existent gauge")
				}
			})

			t.Run("TestStringMethod", func(t *testing.T) {
				str := s.String(context.TODO())
				if len(str) == 0 {
					t.Errorf("String method returned an empty string")
				}
			})
		})
	}
}
