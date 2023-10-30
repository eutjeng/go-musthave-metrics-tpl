package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metricsChan := make(chan []models.Metrics, 10)
	defer close(metricsChan)
	reportChan := make(chan []models.Metrics, 10)
	defer close(reportChan)

	sem := semaphore.NewWeighted(int64(cfg.RateLimit))

	go metrics.GatherStandardMetrics(cfg, sugar, metricsChan)
	go metrics.GatherAdditionalMetrics(cfg, sugar, metricsChan)
	go metrics.DispatchMetrics(ctx, cfg, sugar, client, metricsChan, reportChan, sem)

	select {}
}
