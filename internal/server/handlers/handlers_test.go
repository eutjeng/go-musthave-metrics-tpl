package handlers_test

import (
	"context"
	"io"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandleUpdateAndGetMetrics(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	r := chi.NewRouter()
	sugar := zap.NewExample().Sugar()
	cfg := &config.Config{}

	r.HandleFunc("/update", handlers.HandleUpdateMetric(context.TODO(), cfg, sugar, storage, false))
	r.HandleFunc("/value", handlers.HandleGetMetric(context.TODO(), cfg, sugar, storage))

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
	sugar := zap.NewExample().Sugar()
	cfg := &config.Config{}

	_ = storage.UpdateGauge(context.TODO(), "testGauge", 42.2, false)
	_ = storage.UpdateCounter(context.TODO(), "testCounter", 42, false)

	r := chi.NewRouter()
	r.Get("/", handlers.HandleMetricsHTML(context.TODO(), cfg, sugar, storage))

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
