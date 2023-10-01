package filestorage

import (
	"os"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadFromFile(t *testing.T) {
	// create a temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	// clean up
	defer os.RemoveAll(tempDir)

	// test data
	testCases := []struct {
		counters map[string]int64
		gauges   map[string]float64
		name     string
	}{
		{
			name: "case1",
			gauges: map[string]float64{
				"gauge1": 1.23,
				"gauge2": 4.56,
			},
			counters: map[string]int64{
				"counter1": 1,
				"counter2": 2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// initialize storage
			s := storage.NewInMemoryStorage()
			s.SetMetricsData(tc.gauges, tc.counters)

			// path to temp file
			filePath := tempDir + "/metrics.json"

			// save to file
			if err := SaveToFile(s, filePath); err != nil {
				t.Errorf("failed to save to file: %v", err)
			}

			// load from file
			if err := LoadFromFile(s, filePath); err != nil {
				t.Errorf("failed to load from file: %v", err)
			}

			// validate data
			gauges, counters := s.GetMetricsData()
			assert.Equal(t, tc.gauges, gauges)
			assert.Equal(t, tc.counters, counters)
		})
	}
}
