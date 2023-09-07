package handlers_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandleUpdateAndGetMetrics(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	r := chi.NewRouter()

	r.Post("/update/{type}/{name}/{value}", handlers.HandleUpdateMetric(storage))
	r.Get("/value/{type}/{name}", handlers.HandleGetMetric(storage))

	ts := httptest.NewServer(r)
	defer ts.Close()

	testCases := []struct {
		name           string
		updateUrl      string
		getUrl         string
		expectedStatus int
		expectedValue  string
		skipGet        bool
	}{
		{
			name:           "Valid gauge",
			updateUrl:      "/update/gauge/someTest/42.2",
			getUrl:         "/value/gauge/someTest",
			expectedStatus: http.StatusOK,
			expectedValue:  "42.2",
		},
		{
			name:           "Valid counter",
			updateUrl:      "/update/counter/someTest/42",
			getUrl:         "/value/counter/someTest",
			expectedStatus: http.StatusOK,
			expectedValue:  "42",
		},
		{
			name:           "Invalid update path",
			updateUrl:      "/update/gauge/",
			expectedStatus: http.StatusNotFound,
			skipGet:        true,
		},
		{
			name:           "Invalid update metric type",
			updateUrl:      "/update/invalidType/temperature/23.4",
			expectedStatus: http.StatusNotFound,
			skipGet:        true,
		},
		{
			name:           "Get non-existing gauge",
			getUrl:         "/value/gauge/nonExisting",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Get non-existing counter",
			getUrl:         "/value/counter/nonExisting",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.updateUrl != "" {
				updateReq, err := http.NewRequest("POST", ts.URL+tc.updateUrl, nil)
				assert.NoError(t, err)
				updateResp, err := ts.Client().Do(updateReq)
				assert.NoError(t, err)
				updateResp.Body.Close()
				assert.Equal(t, tc.expectedStatus, updateResp.StatusCode)
			}

			if !tc.skipGet {
				getReq, err := http.NewRequest("GET", ts.URL+tc.getUrl, nil)
				assert.NoError(t, err)
				getResp, err := ts.Client().Do(getReq)
				assert.NoError(t, err)
				defer getResp.Body.Close()

				body, _ := io.ReadAll(getResp.Body)
				assert.Equal(t, tc.expectedStatus, getResp.StatusCode)

				if tc.expectedValue != "" {
					assert.Equal(t, tc.expectedValue, string(body))
				}
			}
		})
	}
}

func TestHandleMetricsHTML(t *testing.T) {
	storage := storage.NewInMemoryStorage()
	_ = storage.UpdateGauge("testGauge", 42.2)
	_ = storage.UpdateCounter("testCounter", 42)

	r := chi.NewRouter()
	r.Get("/", handlers.HandleMetricsHTML(storage))

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
