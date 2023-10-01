package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config represents the configuration options for the server, populated via environment variables
type Config struct {
	Addr            string        `env:"ADDRESS"`           // the address and port on which the server will run
	Environment     string        `env:"ENVIRONMENT"`       // the application's environment, can be 'development' or 'production'
	FileStoragePath string        `env:"FILE_STORAGE_PATH"` // the filename where the current metrics are saved
	Restore         bool          `env:"RESTORE"`           // whether to restore previously saved values from a file upon server startup
	ReportInterval  time.Duration `env:"REPORT_INTERVAL"`   // interval for sending metrics to the server, in seconds
	PollInterval    time.Duration `env:"POLL_INTERVAL"`     // interval for polling metrics from the runtime package, in seconds
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`    // time interval for saving the current metrics to disk, in seconds
	ReadTimeout     time.Duration `env:"READ_TIMEOUT"`      // read timeout for the server, in seconds
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT"`     // write timeout for the server, in seconds
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT"`      // idle timeout for server connections, in seconds
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

	if !isValidEnvironment(cfg.Environment) {
		return fmt.Errorf("invalid environment: %s. Possible values are 'development' or 'production'", cfg.Environment)
	}

	return nil
}

// loadFromFlags populates Config fields based on command-line flags
func loadFromFlags(cfg *Config) error {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	addr := flagSet.String("a", defaultAddr, "Specify the address and port on which the server will run")
	env := flagSet.String("e", defaultEnvironment, "Specify the application's environment. Possible values are 'development' or 'production'")
	fileStoragePath := flagSet.String("f", defaultFileStoragePath, "Specify the filename where current metric values will be saved")
	restore := flagSet.Bool("rs", defaultRestore, "Enable or disable the restoration of previously saved values from a file upon server startup")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "Set the interval for polling metrics from the runtime package, in seconds")
	reportInterval := flagSet.Int64("r", defaultReportInterval, "Set the interval for sending metrics to the server, in seconds")
	storeInterval := flagSet.Int64("i", defaultStoreInterval, "Set the interval for saving the current server metrics to disk, in seconds")
	readTimeout := flagSet.Int64("rt", defaultReadTimeout, "Specify the read timeout for the server, in seconds")
	writeTimeout := flagSet.Int64("wt", defaultWriteTimeout, "Specify the write timeout for the server, in seconds")
	idleTimeout := flagSet.Int64("it", defaultIdleTimeout, "Specify the idle timeout for server connections, in seconds")

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

	if !isValidEnvironment(cfg.Environment) {
		return fmt.Errorf("invalid environment: %s. Possible values are 'development' or 'production'", cfg.Environment)
	}

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

func isValidEnvironment(env string) bool {
	return env == "development" || env == "production"
}
