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
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/hash"
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

func compressData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func sendRequestWithHashing(cfg *config.Config, sugar *zap.SugaredLogger, client *resty.Client, url string, compressedBody []byte, hash string) error {
	sugar.Infof("request hash: %s", hash)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("HashSHA256", hash).
		SetBody(compressedBody).
		Post(url)

	if err != nil {
		sugar.Errorw("error sending request for metric", "error", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		sugar.Errorw("received non-OK response for metric", "status", resp.Status())
		return fmt.Errorf("received non-OK response: %s", resp.Status())
	}

	return nil
}

func reportMetrics(cfg *config.Config, sugar *zap.SugaredLogger, url string, client *resty.Client, res []models.Metrics) {
	jsonData, err := json.Marshal(res)
	if err != nil {
		sugar.Errorw("json marshaling failed", "error", err)
		return
	}

	hash := hash.ComputeHash(jsonData, cfg.Key)

	compressedData, err := compressData(jsonData)
	if err != nil {
		sugar.Errorw("failed to compress json data", "error", err)
		return
	}

	if err := sendRequestWithHashing(cfg, sugar, client, url, compressedData, hash); err != nil {
		sugar.Errorw("failed to send request with hashing", "error", err)
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

	reportMetrics(cfg, sugar, url, client, response)
}
