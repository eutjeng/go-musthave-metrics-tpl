package main

import (
	"log"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/metrics"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/appinit"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/semaphore"
)

func main() {
	client := resty.New()
	cfg, sugar, syncFunc, err := appinit.InitAgentApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %s", err)
	}
	defer syncFunc()

	metricsChan := make(chan []models.Metrics, 10)
	defer close(metricsChan)

	sem := semaphore.NewWeighted(int64(cfg.RateLimit))
	sugar.Infof("Rate limit value: %v", cfg.RateLimit)

	go metrics.GatherStandardMetrics(cfg, sugar, metricsChan)
	go metrics.GatherAdditionalMetrics(cfg, sugar, metricsChan)
	go metrics.DispatchMetrics(cfg, sugar, client, metricsChan, sem)

	select {}
}
