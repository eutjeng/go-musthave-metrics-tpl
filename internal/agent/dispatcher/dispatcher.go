package dispatcher

import (
	"context"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/reporter"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/agent/utils"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/config"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

// reportMetricsAsync asynchronously reports metrics.
func reportMetricsAsync(metrics []models.Metrics, cfg *config.Config, sugar *zap.SugaredLogger, client *resty.Client, sem *semaphore.Weighted) {
	defer func() {
		sem.Release(1)
	}()
	url := utils.GenerateMetricURL(cfg.Addr)
	err := reporter.ReportMetrics(cfg, sugar, url, client, metrics)
	if err != nil {
		sugar.Errorf("Failed to report metrics: %v", err)
	} else {
		sugar.Infof("Metrics reported successfully")
	}
}

// DispatchMetrics is a function responsible for handling the metrics collected by other routines.
// It receives metrics through a channel, aggregates them, and dispatches them for reporting at regular intervals.
// The function takes a context for cancellation, a configuration object, a logger, a REST client,
// a channel for incoming metrics, a channel for reporting metrics, and a semaphore for controlling concurrency.
// It will stop dispatching metrics and exit if the context is cancelled.
func DispatchMetrics(ctx context.Context, cfg *config.Config, sugar *zap.SugaredLogger, client *resty.Client, ch chan []models.Metrics, reportCh chan []models.Metrics, sem *semaphore.Weighted) {
	var aggregateMetrics []models.Metrics

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case metrics := <-ch:
			sugar.Infof("Received metrics: %v", metrics)
			aggregateMetrics = append(aggregateMetrics, metrics...)

		case <-reportTicker.C:
			if len(aggregateMetrics) == 0 {
				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				sugar.Errorf("Failed to acquire semaphore: %v", err)
				continue
			}

			go reportMetricsAsync(aggregateMetrics, cfg, sugar, client, sem)

			aggregateMetrics = nil

		case <-ctx.Done():
			sugar.Infof("Context cancelled, exiting DispatchMetrics")
			return
		}
	}
}
