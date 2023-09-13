package reporter

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/config"
	"github.com/go-resty/resty/v2"
)

func reportMetric(metricType, name string, value interface{}, client *resty.Client) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", config.ServerAddress, metricType, name, value)
	resp, err := client.R().Post(url)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		fmt.Println("Received non-OK response:", resp.Status())
	}
}

func reportGauge(name string, value float64, client *resty.Client) {
	reportMetric("gauge", name, value, client)
}

func reportCounter(name string, value int64, client *resty.Client) {
	reportMetric("counter", name, value, client)
}

func ReportMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	client := resty.New()

	reportGauge("Alloc", float64(m.Alloc), client)
	reportGauge("BuckHashSys", float64(m.BuckHashSys), client)
	reportGauge("Frees", float64(m.Frees), client)
	reportGauge("GCCPUFraction", float64(m.GCCPUFraction), client)
	reportGauge("GCSys", float64(m.GCSys), client)
	reportGauge("HeapAlloc", float64(m.HeapAlloc), client)
	reportGauge("HeapIdle", float64(m.HeapIdle), client)
	reportGauge("HeapInuse", float64(m.HeapInuse), client)
	reportGauge("HeapObjects", float64(m.HeapObjects), client)
	reportGauge("HeapReleased", float64(m.HeapReleased), client)
	reportGauge("HeapSys", float64(m.HeapSys), client)
	reportGauge("LastGC", float64(m.LastGC), client)
	reportGauge("Lookups", float64(m.Lookups), client)
	reportGauge("MCacheInuse", float64(m.MCacheInuse), client)
	reportGauge("MCacheSys", float64(m.MCacheSys), client)
	reportGauge("MSpanInuse", float64(m.MSpanInuse), client)
	reportGauge("MSpanSys", float64(m.MSpanSys), client)
	reportGauge("Mallocs", float64(m.Mallocs), client)
	reportGauge("NextGC", float64(m.NextGC), client)
	reportGauge("NumForcedGC", float64(m.NumForcedGC), client)
	reportGauge("NumGC", float64(m.NumGC), client)
	reportGauge("OtherSys", float64(m.OtherSys), client)
	reportGauge("PauseTotalNs", float64(m.PauseTotalNs), client)
	reportGauge("StackInuse", float64(m.StackInuse), client)
	reportGauge("StackSys", float64(m.StackSys), client)
	reportGauge("Sys", float64(m.Sys), client)
	reportGauge("TotalAlloc", float64(m.TotalAlloc), client)

	reportGauge("RandomValue", config.RandomValue, client)
	reportCounter("PollCount", config.PollCount, client)
}
