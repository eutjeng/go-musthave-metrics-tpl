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
	Environment    string
	ReportInterval time.Duration
	PollInterval   time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
}

const (
	defaultAddr           = ":8080"
	defaultEnvironment    = "development"
	defaultReportInterval = 10 // in seconds
	defaultPollInterval   = 2  // in seconds
	defaultReadTimeout    = 5  // in seconds
	defaultWriteTimeout   = 10 // in seconds
	defaultIdleTimeout    = 15 // in seconds
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
		Environment    string `env:"ENVIRONMENT"`
		ReportInterval int64  `env:"REPORT_INTERVAL"`
		PollInterval   int64  `env:"POLL_INTERVAL"`
		ReadTimeout    int64  `env:"READ_TIMEOUT"`
		WriteTimeout   int64  `env:"WRITE_TIMEOUT"`
		IdleTimeout    int64  `env:"IDLE_TIMEOUT"`
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
	if tempCfg.Environment != "" {
		cfg.Environment = tempCfg.Environment
	}

	if tempCfg.ReadTimeout > 0 {
		cfg.ReadTimeout = time.Duration(tempCfg.ReadTimeout) * time.Second
	}
	if tempCfg.WriteTimeout > 0 {
		cfg.WriteTimeout = time.Duration(tempCfg.WriteTimeout) * time.Second
	}
	if tempCfg.IdleTimeout > 0 {
		cfg.IdleTimeout = time.Duration(tempCfg.IdleTimeout) * time.Second
	}

	return nil
}

// loadFromFlags populates Config fields based on command-line flags
func loadFromFlags(cfg *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	addr := flagSet.String("a", defaultAddr, "address and port to run server")
	reportInterval := flagSet.Int64("r", defaultReportInterval, "frequency of sending metrics to the server (seconds)")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "frequency of metrics polling from the runtime package (seconds)")
	env := flagSet.String("e", defaultEnvironment, "application environment (development|production)")
	readTimeout := flagSet.Int64("rt", defaultReadTimeout, "read timeout in seconds")
	writeTimeout := flagSet.Int64("wt", defaultWriteTimeout, "write timeout in seconds")
	idleTimeout := flagSet.Int64("it", defaultIdleTimeout, "idle timeout in seconds")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg.Addr = *addr
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.Environment = *env
	cfg.ReadTimeout = time.Duration(*readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(*writeTimeout) * time.Second
	cfg.IdleTimeout = time.Duration(*idleTimeout) * time.Second

	return nil
}
