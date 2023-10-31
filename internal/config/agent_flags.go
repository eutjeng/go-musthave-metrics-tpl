package config

import (
	"flag"
	"time"
)

// loadAgentFlags loads flags related to agent settings
func loadAgentFlags(flagSet *flag.FlagSet, cfg *Config) PostParseSetter {
	reportInterval := flagSet.Int64("r", defaultReportInterval, "Set the interval for sending metrics to the server, in seconds")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "Set the interval for polling metrics from the runtime package, in seconds")
	rateLimit := flagSet.Int("l", defaultRateLimit, "Specify the rate limit for outgoing requests. This is the maximum number of concurrent outgoing requests allowed.")

	return func(cfg *Config) {
		cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
		cfg.PollInterval = time.Duration(*pollInterval) * time.Second
		cfg.RateLimit = *rateLimit

	}
}
