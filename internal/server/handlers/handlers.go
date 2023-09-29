package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"go.uber.org/zap"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/constants"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
)

func HandleUpdateMetric(sugar *zap.SugaredLogger, storage storage.MetricStorage) http.HandlerFunc {
	var err error

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sugar.Errorw("Got request with bad method", zap.String("method", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		dec := json.NewDecoder(r.Body)
		var req models.Metrics

		if decodeErr := dec.Decode(&req); decodeErr != nil {
			sugar.Errorw("Cannot decode request JSON body", decodeErr)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		switch req.MType {
		case constants.MetricTypeGauge:
			if req.Value != nil {
				err = storage.UpdateGauge(req.ID, *req.Value)
			} else {
				http.Error(w, "Missing 'value' for gauge", http.StatusBadRequest)
			}
		case constants.MetricTypeCounter:
			if req.Delta != nil {
				err = storage.UpdateCounter(req.ID, *req.Delta)
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(req); err != nil {
			sugar.Errorw("Cannot encode response JSON body", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func HandleGetMetric(sugar *zap.SugaredLogger, storage storage.MetricStorage) http.HandlerFunc {
	var (
		v   interface{}
		err error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		var req models.MetricsQuery

		if decodeErr := dec.Decode(&req); decodeErr != nil {
			sugar.Errorw("Cannot decode request JSON body", decodeErr)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		switch req.MType {
		case constants.MetricTypeGauge:
			v, err = storage.GetGauge(req.ID)

		case constants.MetricTypeCounter:
			v, err = storage.GetCounter(req.ID)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		resp := models.Metrics{
			ID:    req.ID,
			MType: req.MType,
		}

		if req.MType == constants.MetricTypeGauge {
			if value, ok := v.(float64); ok {
				resp.Value = &value
			} else {
				sugar.Errorw("Unexpected type for gauge value", "received", v)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
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
	}
}

func HandleMetricsHTML(storage fmt.Stringer) http.HandlerFunc {
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
