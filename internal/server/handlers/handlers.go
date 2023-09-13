package handlers

import (
	"fmt"
	"html"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
)

func HandleUpdateMetric(storage storage.MetricStorage) http.HandlerFunc {
	var err error

	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		switch metricType {
		case "gauge":
			if v, e := utils.ParseFloat(metricValue); e == nil {
				err = storage.UpdateGauge(metricName, v)
			} else {
				http.Error(w, "Invalid value type for gauge", http.StatusBadRequest)
			}
		case "counter":
			if v, e := utils.ParseInt(metricValue); e == nil {
				err = storage.UpdateCounter(metricName, v)
			} else {
				http.Error(w, "Invalid value type for counter", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
		}

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}
}

func HandleGetMetric(storage storage.MetricStorage) http.HandlerFunc {
	var (
		v   interface{}
		err error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		switch metricType {
		case "gauge":
			v, err = storage.GetGauge(metricName)

		case "counter":
			v, err = storage.GetCounter(metricName)

		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}

		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if _, err := io.WriteString(w, fmt.Sprint(v)); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func HandleMetricsHTML(storage storage.MetricStorage) http.HandlerFunc {
	metricsString := storage.String()

	return func(w http.ResponseWriter, r *http.Request) {
		html := "<html><head><title>Metrics</title>" +
			"<style>body { background-color: black; color: white; font-size: 1.2rem; line-height: 1.5rem }</style>" +
			"</head><body><pre>" +
			html.EscapeString(metricsString) +
			"</pre></body></html>"

		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(html)); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
