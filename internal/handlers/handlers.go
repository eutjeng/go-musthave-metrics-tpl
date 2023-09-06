package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/storage"
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
			var value float64
			value, err = utils.ParseFloat(mV)

			if err == nil {
				err = storage.UpdateGauge(mN, value)
			}

		case "counter":
			var value int64
			value, err = utils.ParseInt(mV)

			if err == nil {
				err = storage.UpdateCounter(mN, value)
			}

		default:
			err = fmt.Errorf("invalid metric type")
		}

		if err != nil {
			fmt.Println("Error:", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

	}
}
