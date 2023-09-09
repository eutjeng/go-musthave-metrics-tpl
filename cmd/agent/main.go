package main

import (
	"log"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/reporter"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/updater"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
)

func main() {
	var pollCount int64
	var randomValue float64

	cfg, err := config.ParseConfig()

	if err != nil {
		log.Fatalf("Error while parsing config: %s", err)
		return
	}

	go func() {
		for {
			updater.UpdateMetrics(&pollCount, &randomValue)
			time.Sleep(cfg.PollInterval)
		}
	}()

	for {
		reporter.ReportMetrics(cfg, randomValue, pollCount)
		time.Sleep(cfg.ReportInterval)
	}
}
