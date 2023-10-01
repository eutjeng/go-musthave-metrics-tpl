package main

import (
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/metrics"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/appinit"
	"github.com/go-resty/resty/v2"
)

func main() {
	var pollCount int64
	var randomValue float64

	client := resty.New()
	cfg, sugar, syncFunc := appinit.InitApp()
	defer syncFunc()

	go func() {
		for {
			metrics.ReportMetrics(sugar, cfg, client, randomValue, pollCount)
			time.Sleep(cfg.ReportInterval)

		}
	}()

	for {
		metrics.UpdateMetrics(&pollCount, &randomValue)
		time.Sleep(cfg.PollInterval)
	}
}
