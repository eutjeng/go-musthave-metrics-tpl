package main

import (
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/reporter"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/updater"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
)

func main() {
	config.ParseFlags()

	go func() {
		for {
			updater.UpdateMetrics()
			time.Sleep(config.PollInterval)
		}
	}()

	for {
		reporter.ReportMetrics()
		time.Sleep(config.ReportInterval)
	}
}
