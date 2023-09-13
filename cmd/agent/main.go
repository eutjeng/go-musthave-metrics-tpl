package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	serverAddress  = "http://localhost:8080"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

var (
	pollCount   int64
	randomValue float64
)

func main() {
	go func() {
		for {
			updateMetrics()
			time.Sleep(pollInterval)
		}
	}()

	for {
		reportMetrics()
		time.Sleep(reportInterval)
	}
}

func updateMetrics() {
	pollCount++
	randomValue = rand.Float64()
}

func reportMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	reportGauge("Alloc", float64(m.Alloc))
	reportGauge("BuckHashSys", float64(m.BuckHashSys))
	reportGauge("Frees", float64(m.Frees))
	reportGauge("GCCPUFraction", float64(m.GCCPUFraction))
	reportGauge("GCSys", float64(m.GCSys))
	reportGauge("HeapAlloc", float64(m.HeapAlloc))
	reportGauge("HeapIdle", float64(m.HeapIdle))
	reportGauge("HeapInuse", float64(m.HeapInuse))
	reportGauge("HeapObjects", float64(m.HeapObjects))
	reportGauge("HeapReleased", float64(m.HeapReleased))
	reportGauge("HeapSys", float64(m.HeapSys))
	reportGauge("LastGC", float64(m.LastGC))
	reportGauge("Lookups", float64(m.Lookups))
	reportGauge("MCacheInuse", float64(m.MCacheInuse))
	reportGauge("MCacheSys", float64(m.MCacheSys))
	reportGauge("MSpanInuse", float64(m.MSpanInuse))
	reportGauge("MSpanSys", float64(m.MSpanSys))
	reportGauge("Mallocs", float64(m.Mallocs))
	reportGauge("NextGC", float64(m.NextGC))
	reportGauge("NumForcedGC", float64(m.NumForcedGC))
	reportGauge("NumGC", float64(m.NumGC))
	reportGauge("OtherSys", float64(m.OtherSys))
	reportGauge("PauseTotalNs", float64(m.PauseTotalNs))
	reportGauge("StackInuse", float64(m.StackInuse))
	reportGauge("StackSys", float64(m.StackSys))
	reportGauge("Sys", float64(m.Sys))
	reportGauge("TotalAlloc", float64(m.TotalAlloc))

	reportGauge("RandomValue", randomValue)
	reportCounter("PollCount", pollCount)
}

func reportMetric(metricType, name string, value interface{}) {
	url := fmt.Sprintf("%s/update/%s/%s/%v", serverAddress, metricType, name, value)
	resp, err := http.Post(url, "text/plain", nil)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	resp.Body.Close()
}

func reportGauge(name string, value float64) {
	reportMetric("gauge", name, value)
}

func reportCounter(name string, value int64) {
	reportMetric("counter", name, value)
}
