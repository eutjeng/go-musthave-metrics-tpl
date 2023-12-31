package metrics

import (
	"bytes"
	"compress/gzip"
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

const urlTemplate = "%s/updates"

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

func reportMetrics(sugar *zap.SugaredLogger, url string, client *resty.Client, res []models.Metrics) {
	jsonData, err := json.Marshal(res)

	if err != nil {
		sugar.Errorw("JSON marshaling failed", err)
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, gzErr := gz.Write(jsonData); gzErr != nil {
		sugar.Errorw("Failed to write gzipped JSON data", err)
		return
	}
	if gzErr := gz.Close(); gzErr != nil {
		sugar.Errorw("Failed to close gzip writer", err)
		return
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(b.Bytes()).
		Post(url)

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
	url := generateMetricURL(cfg.Addr)

	gauges := collectMemoryMetrics()
	counters := map[string]int64{
		"PollCount": pollCount,
	}

	var response []models.Metrics

	gauges["RandomValue"] = randomValue

	for name, value := range gauges {
		localValue := value
		metric := models.Metrics{
			ID:    name,
			MType: constants.MetricTypeGauge,
			Value: &localValue,
		}

		response = append(response, metric)
	}

	for name, delta := range counters {
		localDelta := delta
		metric := models.Metrics{
			ID:    name,
			MType: constants.MetricTypeCounter,
			Delta: &localDelta,
		}
		response = append(response, metric)
	}

	reportMetrics(sugar, url, client, response)
}
