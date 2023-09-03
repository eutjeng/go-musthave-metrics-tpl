package handlers

import (
	"fmt"
	"net/http"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/storage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
)

func HandleUpdateMetrics(storage storage.MetricStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := utils.SplitPath(r.URL.Path)

		if len(parts) != 4 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		metricType, metricName, metricValue := parts[1], parts[2], parts[3]
		var err error

		switch metricType {
		case "gauge":
			var value float64
			value, err = utils.ParseFloat(metricValue)

			if err == nil {
				err = storage.UpdateGauge(metricName, value)
			}

		case "counter":
			var value int64
			value, err = utils.ParseInt(metricValue)

			if err == nil {
				err = storage.UpdateCounter(metricName, value)
			}

		default:
			err = fmt.Errorf("Invalid metric type")
		}

		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

	}
}
