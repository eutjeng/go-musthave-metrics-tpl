package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr            string        `env:"ADDRESS"`
	Environment     string        `env:"ENVIRONMENT"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `eng:"RESTORE"`
	ReportInterval  time.Duration `env:"REPORT_INTERVAL"`
	PollInterval    time.Duration `env:"POLL_INTERVAL"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT"`
}

const (
	defaultAddr            = ":8080"
	defaultEnvironment     = "development"
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true
	defaultReportInterval  = 10  // in seconds
	defaultPollInterval    = 2   // in seconds
	defaultReadTimeout     = 5   // in seconds
	defaultWriteTimeout    = 10  // in seconds
	defaultIdleTimeout     = 15  // in seconds
	defaultStoreInterval   = 300 // in seconds
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
	if err := parseEnvWithDuration(cfg); err != nil {
		return err
	}

	return nil
}

// loadFromFlags populates Config fields based on command-line flags
func loadFromFlags(cfg *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	addr := flagSet.String("a", defaultAddr, "address and port to run server (string)")
	reportInterval := flagSet.Int64("r", defaultReportInterval, "frequency of sending metrics to the server (seconds)")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "frequency of metrics polling from the runtime package (seconds)")
	env := flagSet.String("e", defaultEnvironment, "application environment (development|production)")
	restore := flagSet.Bool("rs", defaultRestore, "load previously saved values from the specified file at server startup (bool)")
	fileStoragePath := flagSet.String("f", defaultFileStoragePath, "file name where the current values are saved (string)")
	storeInterval := flagSet.Int64("i", defaultStoreInterval, "time interval after which the current server readings are saved to disk (seconds)")
	readTimeout := flagSet.Int64("rt", defaultReadTimeout, "read timeout (seconds)")
	writeTimeout := flagSet.Int64("wt", defaultWriteTimeout, "write timeout (seconds)")
	idleTimeout := flagSet.Int64("it", defaultIdleTimeout, "idle timeout (seconds)")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg.Addr = *addr
	cfg.Environment = *env
	cfg.FileStoragePath = *fileStoragePath
	cfg.Restore = *restore
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReadTimeout = time.Duration(*readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(*writeTimeout) * time.Second
	cfg.IdleTimeout = time.Duration(*idleTimeout) * time.Second
	cfg.StoreInterval = time.Duration(*storeInterval) * time.Second

	return nil
}

// parseEnvWithDuration fetches time.Duration fields from the environment variables,
// sets them properly with a 's' suffix, and then parses all environment variables
// into the Config struct.
func parseEnvWithDuration(cfg *Config) error {
	envVars := getDurationFields(cfg)
	if err := setEnvVars(envVars); err != nil {
		return err
	}

	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return nil
}

// setEnvVars sets the values of environment variables based on the given map.
func setEnvVars(envVars map[string]string) error {
	for key, val := range envVars {
		if err := os.Setenv(key, val); err != nil {
			return err
		}
	}
	return nil
}

// getDurationFields scans the fields of the given Config struct, extracts
// environment variable names for fields of type time.Duration, and retrieves
// their current values. It returns a map of environment variable names to values.
func getDurationFields(cfg *Config) map[string]string {
	t := reflect.TypeOf(*cfg)
	envVars := make(map[string]string)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type == reflect.TypeOf(time.Duration(0)) {
			envName := field.Tag.Get("env")
			if envName != "" {
				envValue := os.Getenv(envName)
				if envValue != "" {
					envVars[envName] = envValue + "s"
				}
			}
		}
	}

	return envVars
}
