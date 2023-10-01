package filestorage

import (
	"fmt"
	"os"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveAndLoad performs both SaveToFile and LoadFromFile, and returns any occurring error
func saveAndLoad(s *storage.InMemoryStorage, filePath string) error {
	if err := SaveToFile(s, filePath); err != nil {
		return err
	}
	return LoadFromFile(s, filePath)
}

func TestSaveAndLoadFromFile(t *testing.T) {
	// Arrange: create a temporary directory
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)

	// clean up
	defer func() {
		removeErr := os.RemoveAll(tempDir)
		assert.NoError(t, removeErr)
	}()

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
		{
			name: "caseWithZeroValues",
			gauges: map[string]float64{
				"gauge1": 0.0,
			},
			counters: map[string]int64{
				"counter1": 0,
			},
		},
		{
			name: "caseWithNegativeValues",
			gauges: map[string]float64{
				"gauge1": -1.23,
			},
			counters: map[string]int64{
				"counter1": -1,
			},
		},
		{
			name:     "caseWithEmptyMaps",
			gauges:   map[string]float64{},
			counters: map[string]int64{},
		},
		{
			name:     "caseWithNilMaps",
			gauges:   nil,
			counters: nil,
		},
		{
			name: "caseWithLargeFile",
			gauges: func() map[string]float64 {
				m := make(map[string]float64)
				for i := 0; i < 10000; i++ {
					m[fmt.Sprintf("gauge%d", i)] = float64(i)
				}
				return m
			}(),
			counters: func() map[string]int64 {
				m := make(map[string]int64)
				for i := 0; i < 10000; i++ {
					m[fmt.Sprintf("counter%d", i)] = int64(i)
				}
				return m
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: initialize storage
			s := storage.NewInMemoryStorage()
			s.SetMetricsData(tc.gauges, tc.counters)

			// path to temp file
			filePath := tempDir + "/metrics.json"

			// Act: save to file and load from it
			err := saveAndLoad(s, filePath)
			require.NoError(t, err)

			// Assert: validate data
			gauges, counters := s.GetMetricsData()
			assert.Equal(t, tc.gauges, gauges)
			assert.Equal(t, tc.counters, counters)
		})
	}
}
