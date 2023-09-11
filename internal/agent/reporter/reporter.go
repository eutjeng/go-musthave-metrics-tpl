package reporter

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
	"github.com/go-resty/resty/v2"
)

const (
	URLTemplate       = "%s/update/%s/%s/%v"
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
)

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
		"OtherSys":      float64(m.OtherSys),
		"NumGC":         float64(m.NumGC),
		"StackInuse":    float64(m.StackInuse),
		"StackSys":      float64(m.StackSys),
		"Sys":           float64(m.Sys),
		"TotalAlloc":    float64(m.TotalAlloc),
	}
}

func reportSingleMetric(url string, client *resty.Client) {
	resp, err := client.R().Post(url)

	if err != nil {
		log.Printf("Error sending request for metric: %v\n", err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Received non-OK response for metric: %s\n", resp.Status())
	}
}

func generateMetricURL(addr, metricType, name string, value interface{}) string {
	return fmt.Sprintf(URLTemplate, utils.EnsureHTTPScheme(addr), metricType, name, value)
}

func ReportMetrics(cfg *config.Config, client *resty.Client, RandomValue float64, PollCount int64) {
	gauges := collectMemoryMetrics()
	gauges["RandomValue"] = RandomValue

	counters := map[string]int64{
		"PollCount": PollCount,
	}

	for name, value := range gauges {
		url := generateMetricURL(cfg.Addr, GaugeMetricType, name, value)
		reportSingleMetric(url, client)
	}

	for name, value := range counters {
		url := generateMetricURL(cfg.Addr, CounterMetricType, name, value)
		reportSingleMetric(url, client)
	}
}
