package metrics

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/constants"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const urlTemplate = "%s/update"

func UpdateMetrics(pollCount *int64, randomValue *float64) {
	*pollCount++
	*randomValue = rand.Float64()
}

func collectMemoryMetrics() map[string]float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]float64{
		"Alloc":         float64(m.Alloc),
		"BuckHashSys":   float64(m.BuckHashSys),
		"Frees":         float64(m.Frees),
		"GCCPUFraction": float64(m.GCCPUFraction),
		"GCSys":         float64(m.GCSys),
		"HeapAlloc":     float64(m.HeapAlloc),
		"HeapIdle":      float64(m.HeapIdle),
		"HeapInuse":     float64(m.HeapInuse),
		"HeapObjects":   float64(m.HeapObjects),
		"HeapReleased":  float64(m.HeapReleased),
		"HeapSys":       float64(m.HeapSys),
		"LastGC":        float64(m.LastGC),
		"Lookups":       float64(m.Lookups),
		"MCacheInuse":   float64(m.MCacheInuse),
		"MCacheSys":     float64(m.MCacheSys),
		"MSpanInuse":    float64(m.MSpanInuse),
		"MSpanSys":      float64(m.MSpanSys),
		"Mallocs":       float64(m.Mallocs),
		"NextGC":        float64(m.NextGC),
		"NumForcedGC":   float64(m.NumForcedGC),
		"OtherSys":      float64(m.OtherSys),
		"NumGC":         float64(m.NumGC),
		"PauseTotalNs":  float64(m.PauseTotalNs),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
	}
}

func reportSingleMetric(sugar *zap.SugaredLogger, url string, client *resty.Client, res *models.Metrics) {
	jsonData, err := json.Marshal(res)

	if err != nil {
		sugar.Errorw("JSON marshaling failed", err)
	}

	resp, err := client.R().SetHeader("Content-Type", "application/json").SetBody(jsonData).Post(url)

	if err != nil {
		sugar.Errorw("Error sending request for metric", err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		sugar.Errorw("Received non-OK response for metric", resp.Status())
	}
}

func generateMetricURL(addr string) string {
	return fmt.Sprintf(urlTemplate, utils.EnsureHTTPScheme(addr))
}

func ReportMetrics(sugar *zap.SugaredLogger, cfg *config.Config, client *resty.Client, randomValue float64, pollCount int64) {
	gauges := collectMemoryMetrics()
	gauges["RandomValue"] = randomValue
	url := generateMetricURL(cfg.Addr)

	counters := map[string]int64{
		"PollCount": pollCount,
	}

	for name, value := range gauges {
		res := models.Metrics{
			ID:    name,
			MType: constants.MetricTypeGauge,
			Value: &value,
		}

		reportSingleMetric(sugar, url, client, &res)
	}

	for name, delta := range counters {
		res := models.Metrics{
			ID:    name,
			MType: constants.MetricTypeCounter,
			Delta: &delta,
		}

		reportSingleMetric(sugar, url, client, &res)
	}
}
