package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr           string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

const (
	defaultAddr           = ":8080"
	defaultReportInterval = 10 // in seconds
	defaultPollInterval   = 2  // in seconds
)

func ParseConfig() (*Config, error) {
	cfg := &Config{}
	if err := loadFromFlags(cfg); err != nil {
		return nil, err
	}
	if err := loadFromEnv(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// loadFromEnv overrides Config fields from environment variables
func loadFromEnv(cfg *Config) error {
	tempCfg := struct {
		Addr           string `env:"ADDRESS"`
		ReportInterval int64  `env:"REPORT_INTERVAL"`
		PollInterval   int64  `env:"POLL_INTERVAL"`
	}{}

	if err := env.Parse(&tempCfg); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if tempCfg.Addr != "" {
		cfg.Addr = tempCfg.Addr
	}
	if tempCfg.ReportInterval > 0 {
		cfg.ReportInterval = time.Duration(tempCfg.ReportInterval) * time.Second
	}
	if tempCfg.PollInterval > 0 {
		cfg.PollInterval = time.Duration(tempCfg.PollInterval) * time.Second
	}

	return nil
}

// loadFromFlags populates Config fields based on command-line flags
func loadFromFlags(cfg *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	addr := flagSet.String("a", defaultAddr, "address and port to run server")
	reportInterval := flagSet.Int64("r", defaultReportInterval, "frequency of sending metrics to the server (seconds)")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "frequency of metrics polling from the runtime package (seconds)")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg.Addr = *addr
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second

	return nil
}
