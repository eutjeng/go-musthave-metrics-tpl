package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestHandleUpdateMetrics(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", tc.url, bytes.NewBuffer(nil))
			assert.NoError(t, err)

			storage := storage.NewInMemoryStorage()
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handlers.HandleUpdateMetrics(storage))

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
