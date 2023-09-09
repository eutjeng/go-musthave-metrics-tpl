package reporter

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/utils"
	"github.com/go-resty/resty/v2"
)

func reportMetric(metricType, name string, value interface{}, client *resty.Client, cfg *config.Config) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", utils.EnsureHTTPScheme(cfg.Addr), metricType, name, value)
	resp, err := client.R().Post(url)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		fmt.Println("Received non-OK response:", resp.Status())
	}
}

func reportGauge(name string, value float64, client *resty.Client, cfg *config.Config) {
	reportMetric("gauge", name, value, client, cfg)
}

func reportCounter(name string, value int64, client *resty.Client, cfg *config.Config) {
	reportMetric("counter", name, value, client, cfg)
}

func ReportMetrics(cfg *config.Config, RandomValue float64, PollCount int64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	client := resty.New()

	reportGauge("Alloc", float64(m.Alloc), client, cfg)
	reportGauge("BuckHashSys", float64(m.BuckHashSys), client, cfg)
	reportGauge("Frees", float64(m.Frees), client, cfg)
	reportGauge("GCCPUFraction", float64(m.GCCPUFraction), client, cfg)
	reportGauge("GCSys", float64(m.GCSys), client, cfg)
	reportGauge("HeapAlloc", float64(m.HeapAlloc), client, cfg)
	reportGauge("HeapIdle", float64(m.HeapIdle), client, cfg)
	reportGauge("HeapInuse", float64(m.HeapInuse), client, cfg)
	reportGauge("HeapObjects", float64(m.HeapObjects), client, cfg)
	reportGauge("HeapReleased", float64(m.HeapReleased), client, cfg)
	reportGauge("HeapSys", float64(m.HeapSys), client, cfg)
	reportGauge("LastGC", float64(m.LastGC), client, cfg)
	reportGauge("Lookups", float64(m.Lookups), client, cfg)
	reportGauge("MCacheInuse", float64(m.MCacheInuse), client, cfg)
	reportGauge("MCacheSys", float64(m.MCacheSys), client, cfg)
	reportGauge("MSpanInuse", float64(m.MSpanInuse), client, cfg)
	reportGauge("MSpanSys", float64(m.MSpanSys), client, cfg)
	reportGauge("Mallocs", float64(m.Mallocs), client, cfg)
	reportGauge("NextGC", float64(m.NextGC), client, cfg)
	reportGauge("NumForcedGC", float64(m.NumForcedGC), client, cfg)
	reportGauge("NumGC", float64(m.NumGC), client, cfg)
	reportGauge("OtherSys", float64(m.OtherSys), client, cfg)
	reportGauge("PauseTotalNs", float64(m.PauseTotalNs), client, cfg)
	reportGauge("StackInuse", float64(m.StackInuse), client, cfg)
	reportGauge("StackSys", float64(m.StackSys), client, cfg)
	reportGauge("Sys", float64(m.Sys), client, cfg)
	reportGauge("TotalAlloc", float64(m.TotalAlloc), client, cfg)

	reportGauge("RandomValue", RandomValue, client, cfg)
	reportCounter("PollCount", PollCount, client, cfg)
}
