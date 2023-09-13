package main

import (
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/reporter"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/updater"
)

func main() {
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
