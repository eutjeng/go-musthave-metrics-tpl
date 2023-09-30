package handlers_test

import (
	"io"
	"time"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockConfig = &config.Config{
	Addr:           ":8080",
	Environment:    "test",
	ReportInterval: time.Second * 10,
	PollInterval:   time.Second * 2,
}

func TestHandleUpdateAndGetMetrics(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	r := chi.NewRouter()

	sugar, _, _ := logger.InitLogger(mockConfig)

	r.HandleFunc("/update", handlers.HandleUpdateMetric(sugar, storage))
	r.HandleFunc("/value", handlers.HandleGetMetric(sugar, storage))

	ts := httptest.NewServer(r)
	defer ts.Close()

	testCases := []struct {
		name           string
		updateBody     string
		getBody        string
		expectedValue  string
		expectedStatus int
	}{
		{
			name:           "Valid gauge",
			updateBody:     `{"id":"someTest","type":"gauge","value":42.2}`,
			getBody:        `{"id":"someTest","type":"gauge"}`,
			expectedStatus: http.StatusOK,
			expectedValue:  `{"id":"someTest","type":"gauge","value":42.2}`,
		},
		{
			name:           "Valid counter",
			updateBody:     `{"id":"someTest","type":"counter","delta":42}`,
			getBody:        `{"id":"someTest","type":"counter"}`,
			expectedStatus: http.StatusOK,
			expectedValue:  `{"id":"someTest","type":"counter","delta":42}`,
		},
		{
			name:           "Invalid update counter type",
			updateBody:     `{"id":"someTest","type":"counter","value": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid update metric type",
			updateBody:     `{"id":"someTest","type":"invalidType","value": "23.4"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// POST request for update
			updateResp, err := http.Post(ts.URL+"/update", "application/json", strings.NewReader(tc.updateBody))
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, updateResp.StatusCode)
			updateResp.Body.Close()

			// POST request for get
			getResp, err := http.Post(ts.URL+"/value", "application/json", strings.NewReader(tc.getBody))
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, getResp.StatusCode)

			if tc.expectedValue != "" {
				bodyBytes, err := io.ReadAll(getResp.Body)
				require.NoError(t, err)
				getResp.Body.Close()
				require.JSONEq(t, tc.expectedValue, string(bodyBytes))
			}
		})
	}
}

func TestHandleMetricsHTML(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	sugar, _, _ := logger.InitLogger(mockConfig)

	_ = storage.UpdateGauge("testGauge", 42.2)
	_ = storage.UpdateCounter("testCounter", 42)

	r := chi.NewRouter()
	r.Get("/", handlers.HandleMetricsHTML(sugar, storage))

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Contains(t, string(body), "testGauge: 42.2")
	assert.Contains(t, string(body), "testCounter: 42")
}
