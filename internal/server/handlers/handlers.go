package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/constants"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
	"github.com/go-chi/chi/v5"
)

// extractMetrics takes an HTTP request and returns metric details extracted from it
// it handles both JSON and URL parameter formats to retrieve metric type, name, value, and delta
// returns an error if unable to decode the request body or URL parameters
func extractMetrics(r *http.Request) (string, string, *float64, *int64, error) {
	contentType := r.Header.Get("Content-Type")

	var metricType string
	var metricName string
	var metricValue *float64
	var metricDelta *int64

	if strings.TrimSpace(contentType) == constants.ApplicationJSON {
		dec := json.NewDecoder(r.Body)
		var req models.Metrics

		if err := dec.Decode(&req); err != nil {
			return "", "", nil, nil, err
		}

		metricType = req.MType
		metricName = req.ID
		metricValue = req.Value
		metricDelta = req.Delta

	} else {
		metricType = chi.URLParam(r, "type")
		metricName = chi.URLParam(r, "name")

		valueStr := chi.URLParam(r, "value")

		switch metricType {
		case constants.MetricTypeGauge:
			if value, err := utils.ParseFloat(valueStr); err == nil {
				metricValue = &value
				metricDelta = nil
			}
		case constants.MetricTypeCounter:
			if delta, err := utils.ParseInt(valueStr); err == nil {
				metricValue = nil
				metricDelta = &delta
			}
		}
	}

	return metricType, metricName, metricValue, metricDelta, nil
}

// HandleUpdateMetric is an HTTP handler that updates a metric in the storage
// it extracts metric information from the request and uses it to update the metric in storage
// responds with an HTTP status and, in case of JSON content type, a JSON-encoded response
func HandleUpdateMetric(sugar *zap.SugaredLogger, storage models.GeneralStorageInterface, shouldNotify bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType, metricName, metricValue, metricDelta, err := extractMetrics(r)

		if err != nil {
			sugar.Errorw("Error when extracting metrics", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		switch metricType {
		case constants.MetricTypeGauge:
			if metricValue != nil {
				err = storage.UpdateGauge(metricName, *metricValue, shouldNotify)
			} else {
				http.Error(w, "Missing 'value' for gauge", http.StatusBadRequest)
				return
			}
		case constants.MetricTypeCounter:
			if metricDelta != nil {
				err = storage.UpdateCounter(metricName, *metricDelta, shouldNotify)
			} else {
				http.Error(w, "Missing 'delta' for counter", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Failed to update metric", http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") == constants.ApplicationJSON {
			response := map[string]interface{}{
				"type": metricType,
				"name": metricName,
			}

			if metricValue != nil {
				response["value"] = metricValue
			}

			if metricDelta != nil {
				response["delta"] = metricDelta
			}

			w.Header().Set("Content-Type", constants.ApplicationJSON)
			w.WriteHeader(http.StatusOK)

			if err := json.NewEncoder(w).Encode(response); err != nil {
				sugar.Errorw("Cannot encode response JSON body", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

// HandleGetMetric is an HTTP handler that retrieves a metric from the storage
// it extracts metric information from the request and uses it to fetch the metric from storage
// responds with the metric value in either JSON format or as a plain string based on the request's Content-Type header
func HandleGetMetric(sugar *zap.SugaredLogger, storage models.GeneralStorageInterface) http.HandlerFunc {
	var v interface{}

	return func(w http.ResponseWriter, r *http.Request) {
		metricType, metricName, _, _, err := extractMetrics(r)

		if err != nil {
			sugar.Errorw("Error when extracting metrics", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		switch metricType {
		case constants.MetricTypeGauge:
			v, err = storage.GetGauge(metricName)

		case constants.MetricTypeCounter:
			v, err = storage.GetCounter(metricName)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if r.Header.Get("Content-Type") == constants.ApplicationJSON {
			resp := models.Metrics{
				ID:    metricName,
				MType: metricType,
			}

			switch metricType {
			case constants.MetricTypeGauge:
				if value, ok := v.(float64); ok {
					resp.Value = &value
				} else {
					sugar.Errorw("Unexpected type for gauge value", "received", v)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

			case constants.MetricTypeCounter:
				if value, ok := v.(int64); ok {
					resp.Delta = &value
				} else {
					sugar.Errorw("Unexpected type for counter delta", "received", v)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				sugar.Errorw("Cannot encode response JSON body", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, fmt.Sprint(v)); err != nil {
			sugar.Errorw("Cannot write to response body", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// HandleMetricsHTML is an HTTP handler that generates an HTML page displaying all metrics
// the page is generated based on the metrics data retrieved from the storage
// responds with an HTML page containing the metrics
func HandleMetricsHTML(sugar *zap.SugaredLogger, storage fmt.Stringer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricsString := storage.String()
		html := "<html><head><title>Metrics</title>" +
			"<style>body { background-color: black; color: white; font-size: 1.2rem; line-height: 1.5rem }</style>" +
			"</head><body><pre>" +
			html.EscapeString(metricsString) +
			"</pre></body></html>"

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(html)); err != nil {
			sugar.Errorw("Cannot write HTML to response body", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
