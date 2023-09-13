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
	return func(w http.ResponseWriter, r *http.Request) {
		mT := chi.URLParam(r, "type")
		mN := chi.URLParam(r, "name")
		mV := chi.URLParam(r, "value")

		var err error

		switch mT {
		case "gauge":
			if v, e := utils.ParseFloat(mV); e == nil {
				err = storage.UpdateGauge(mN, v)
			} else {
				http.Error(w, "Invalid value type for gauge", http.StatusBadRequest)
			}
		case "counter":
			if v, e := utils.ParseInt(mV); e == nil {
				err = storage.UpdateCounter(mN, v)
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
	return func(w http.ResponseWriter, r *http.Request) {
		mT := chi.URLParam(r, "type")
		mN := chi.URLParam(r, "name")

		var (
			v   interface{}
			err error
		)

		switch mT {
		case "gauge":
			v, err = storage.GetGauge(mN)

		case "counter":
			v, err = storage.GetCounter(mN)

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
	return func(w http.ResponseWriter, r *http.Request) {
		metricsString := storage.String()

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
