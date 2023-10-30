package collector

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/utils"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

func GatherStandardMetrics(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, ch chan []models.Metrics) {
	var pollCount int64
	var randomValue float64

	for {
		select {
		case <-ctx.Done():
			sugar.Info("context cancelled, exiting GatherStandardMetrics")
			close(ch)
			return
		default:
			utils.UpdateMetrics(&pollCount, &randomValue)
			memoryMetrics := collectMemoryMetrics()
			metrics := createMetrics(sugar, pollCount, randomValue, memoryMetrics)
			sugar.Infof("Collected metrics: %+v", metrics)
			ch <- metrics
			time.Sleep(cfg.PollInterval)
		}
	}
}

func GatherAdditionalMetrics(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, ch chan []models.Metrics) {
	for {
		select {
		case <-ctx.Done():
			sugar.Info("context cancelled, exiting GatherAdditionalMetrics")
			close(ch)
			return
		default:
			metrics := collectAdditionalMetrics()
			sugar.Infof("Collected additional metrics: %+v", metrics)
			ch <- metrics
			time.Sleep(cfg.PollInterval)
		}
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
