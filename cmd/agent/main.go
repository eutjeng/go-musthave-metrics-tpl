package main

import (
	"log"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/metrics"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/logger"
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

	sugar, syncFunc, err := logger.InitLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}
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
