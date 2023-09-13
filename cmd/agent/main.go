package main

import (
	"log"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/metrics"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/go-resty/resty/v2"
)

func main() {
	var pollCount int64
	var randomValue float64

	client := resty.New()
	cfg, err := config.ParseConfig()

	if err != nil {
		log.Fatalf("Error while parsing config: %s", err)
	}

	go func() {
		for {
			metrics.UpdateMetrics(&pollCount, &randomValue)
			time.Sleep(cfg.PollInterval)
		}
	}()

	for {
		metrics.ReportMetrics(cfg, client, randomValue, pollCount)
		time.Sleep(cfg.ReportInterval)
	}
}
