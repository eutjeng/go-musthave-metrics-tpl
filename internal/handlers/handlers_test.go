package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
	name           string
	url            string
	expectedStatus int
}{
	{
		name:           "Valid gauge",
		url:            "/update/gauge/someTest/42.2",
		expectedStatus: http.StatusOK,
	},
	{
		name:           "Valid Counter",
		url:            "/update/counter/someTest/42",
		expectedStatus: http.StatusOK,
	},
	{
		name:           "Invalid path",
		url:            "/update/gauge/",
		expectedStatus: http.StatusNotFound,
	},
	{
		name:           "Invalid metric type",
		url:            "/update/invalidType/temperature/23.4",
		expectedStatus: http.StatusBadRequest,
	},
}

func TestHandleUpdateMetrics(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", handlers.HandleUpdateMetric(storage.NewInMemoryStorage()))

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", ts.URL+tc.url, nil)
			assert.NoError(t, err)

			resp, err := ts.Client().Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
