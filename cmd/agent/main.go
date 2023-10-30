package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/collector"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/dispatcher"
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

	go collector.GatherStandardMetrics(ctx, cfg, sugar, metricsChan)
	go collector.GatherAdditionalMetrics(ctx, cfg, sugar, metricsChan)
	go dispatcher.DispatchMetrics(ctx, cfg, sugar, client, metricsChan, reportChan, sem)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	<-sigChan
	cancel()
}
