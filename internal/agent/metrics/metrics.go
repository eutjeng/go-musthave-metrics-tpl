package metrics

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/hash"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

const urlTemplate = "%s/updates"

func GatherStandardMetrics(cfg *config.Config, sugar *zap.SugaredLogger, ch chan []models.Metrics) {
	var pollCount int64
	var randomValue float64

	for {
		updateMetrics(&pollCount, &randomValue)
		memoryMetrics := collectMemoryMetrics()
		metrics := createMetrics(sugar, pollCount, randomValue, memoryMetrics)
		sugar.Infof("Collected metrics: %+v", metrics)
		ch <- metrics
		time.Sleep(cfg.PollInterval)
	}
}

func GatherAdditionalMetrics(cfg *config.Config, sugar *zap.SugaredLogger, ch chan []models.Metrics) {
	for {
		metrics := collectAdditionalMetrics()
		sugar.Infof("Collected additional metrics: %+v", metrics)
		ch <- metrics
		time.Sleep(cfg.PollInterval)
	}
}

func DispatchMetrics(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, client *resty.Client, ch chan []models.Metrics, reportCh chan []models.Metrics, sem *semaphore.Weighted) {
	var aggregateMetrics []models.Metrics

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case metrics := <-ch:
			sugar.Infof("Received metrics: %v", metrics)
			aggregateMetrics = append(aggregateMetrics, metrics...)

		case <-reportTicker.C:
			if len(aggregateMetrics) == 0 {
				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				sugar.Errorf("Failed to acquire semaphore: %v", err)
				continue
			}

			go func(metrics []models.Metrics) {
				defer func() {
					sem.Release(1)
				}()
				url := generateMetricURL(cfg.Addr)
				err := reportMetrics(cfg, sugar, url, client, metrics)
				if err != nil {
					sugar.Errorf("Failed to report metrics: %v", err)
				} else {
					sugar.Infof("Metrics reported successfully")
				}
			}(aggregateMetrics)

			aggregateMetrics = nil
		}
	}
}

func updateMetrics(pollCount *int64, randomValue *float64) {
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

func reportMetrics(cfg *config.Config, sugar *zap.SugaredLogger, url string, client *resty.Client, res []models.Metrics) error {
	sugar.Infof("Sending metrics to %s: %v", url, res)
	jsonData, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("json marshaling failed: %w", err)
	}

	hash := hash.ComputeHash(jsonData, cfg.Key)

	compressedData, err := compressData(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress json data: %w", err)
	}

	err = sendRequestWithHashing(cfg, sugar, client, url, compressedData, hash)
	if err != nil {
		return fmt.Errorf("failed to send request with hashing: %w", err)
	}

	return nil
}

func createMetric(id string, mType string, value float64, delta int64) *models.Metrics {
	v := value
	d := delta
	return &models.Metrics{
		ID:    id,
		MType: mType,
		Value: &v,
		Delta: &d,
	}
}

func createMetrics(sugar *zap.SugaredLogger, pollCount int64, randomValue float64, memoryMetrics map[string]float64) []models.Metrics {
	var metrics []models.Metrics

	metrics = append(metrics, *createMetric("PollCount", "counter", 0, pollCount))
	metrics = append(metrics, *createMetric("RandomValue", "gauge", randomValue, 0))

	for metricName, metricValue := range memoryMetrics {
		metrics = append(metrics, *createMetric(metricName, "gauge", metricValue, 0))
	}

	return metrics
}

func collectAdditionalMetrics() []models.Metrics {
	var metrics []models.Metrics

	v, err := mem.VirtualMemory()
	if err == nil {
		metrics = append(metrics, *createMetric("TotalMemory", "gauge", float64(v.Total), 0))
		metrics = append(metrics, *createMetric("FreeMemory", "gauge", float64(v.Free), 0))
	}

	cpuUsages, err := cpu.Percent(0, true)
	if err == nil {
		for i, usage := range cpuUsages {
			metrics = append(metrics, *createMetric("CPUutilization"+fmt.Sprint(i), "gauge", usage, 0))
		}
	}

	return metrics
}

func generateMetricURL(addr string) string {
	return fmt.Sprintf(urlTemplate, utils.EnsureHTTPScheme(addr))
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
