package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config represents the configuration options for the server, populated via environment variables.
// Note: Default values cannot be assigned directly in the struct tags for fields of type time.Duration.
// This is because the values from environment variables are often plain numbers, which Go won't automatically
// convert to time.Duration. Instead, the values are manually parsed and converted in functions like
// parseEnvWithDuration to allow for more flexible input, such as '300' being interpreted as '300s'.
type Config struct {
	Addr            string        `env:"ADDRESS"`           // the address and port on which the server will run
	Environment     string        `env:"ENVIRONMENT"`       // the application's environment, can be 'development' or 'production'
	FileStoragePath string        `env:"FILE_STORAGE_PATH"` // the filename where the current metrics are saved
	DBDSN           string        `env:"DATABASE_DSN"`      // the Data Source Name for connecting to the database
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS"`    // max number of open database connections
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS"`    // max number of idle database connections
	Restore         bool          `env:"RESTORE"`           // whether to restore previously saved values from a file upon server startup
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME"` // max lifetime of a database connection, in seconds
	ReportInterval  time.Duration `env:"REPORT_INTERVAL"`   // interval for sending metrics to the server, in seconds
	PollInterval    time.Duration `env:"POLL_INTERVAL"`     // interval for polling metrics from the runtime package, in seconds
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`    // time interval for saving the current metrics to disk, in seconds
	ReadTimeout     time.Duration `env:"READ_TIMEOUT"`      // read timeout for the server, in seconds
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT"`     // write timeout for the server, in seconds
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT"`      // idle timeout for server connections, in seconds
	MaxRetries      int           `env:"MAX_RETRIES"`       // maximum number of retries for the operation
	InitialDelay    time.Duration `env:"INITIAL_DELAY"`     // initial delay before the first retry attempt, in seconds
	DelayIncrement  time.Duration `env:"DELAY_INCREMENT"`   // incremental delay between retries, in seconds
}

type PostParseSetter func(*Config)
type FlagSetter func(*flag.FlagSet, *Config) PostParseSetter

const (
	defaultAddr            = ":8080"
	defaultEnvironmentDev  = "development"
	defaultEnvironmentProd = "production"
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultDBDSN           = ""
	defaultRestore         = true
	defaultReportInterval  = 10  // in seconds
	defaultPollInterval    = 2   // in seconds
	defaultReadTimeout     = 5   // in seconds
	defaultWriteTimeout    = 10  // in seconds
	defaultIdleTimeout     = 15  // in seconds
	defaultStoreInterval   = 300 // in seconds
	defaultMaxOpenConns    = 25  // in seconds
	defaultMaxIdleConns    = 25  // in seconds
	defaultConnMaxLifetime = 300 // in seconds
	defaultInitialDelay    = 1   // in seconds
	defaultMaxRetries      = 3   // in seconds
	defaultDelayIncrement  = 1   // in seconds
)

// loadAndParseFlags is responsible for configuring, parsing, and validating
// command-line flags. It is a generic utility function designed to work with
// various types of configurations (server, agent, etc.).

// It uses a variable number of 'FlagSetter' functions to decouple the process
// of defining which flags are valid from the process of parsing and applying them.
// This makes the function highly extensible, as new sets of flags can be added
// without altering existing code.

// The function also uses a two-phase approach:
//  1. In the first phase, all the 'FlagSetter' functions are called to register flags.
//     Note: Direct assignment like cfg.addr = *flagSet.String("a", ...)
//     is not possible at this stage because flag values are populated only after
//     flagSet.Parse has been called. Therefore, these functions can also return
//     'PostParseSetter' functions for delayed post-parse actions.
//  2. In the second phase, after parsing, any 'PostParseSetter' functions are called.
//     This is useful because some actions can't be performed until all flags are parsed,
//     for example, you might have flags that depend on each other.
func loadAndParseFlags(cfg *Config, setters ...FlagSetter) error {
	// Create a new flag set. This holds the command-line flags and parameters.
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Initialize a slice to hold functions that will be called after flag parsing.
	var postParseSetters []PostParseSetter

	// Loop through each FlagSetter, allowing them to modify the flag set and
	// register any post-parse actions.
	for _, setter := range setters {
		// Call the FlagSetter and store the PostParseSetter.
		postParseSetters = append(postParseSetters, setter(flagSet, cfg))
	}

	// Parse the command-line flags.
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// After parsing, call each PostParseSetter to finalize the config.
	for _, postParseSetter := range postParseSetters {
		postParseSetter(cfg)
	}

	// Validate the environment setting as a final step.
	if !isValidEnvironment(cfg.Environment) {
		return fmt.Errorf(
			"invalid environment: %s. Possible values are '%s' or '%s'",
			cfg.Environment,
			defaultEnvironmentDev,
			defaultEnvironmentProd,
		)

	}

	return nil
}

// loadGeneralFlags loads flags related to general settings
func loadGeneralFlags(flagSet *flag.FlagSet, cfg *Config) PostParseSetter {
	addr := flagSet.String("a", defaultAddr, "Specify the address and port on which the server will run")
	env := flagSet.String(
		"e",
		defaultEnvironmentDev,
		fmt.Sprintf(
			"Specify the application's environment. Possible values are '%s' or '%s'",
			defaultEnvironmentDev,
			defaultEnvironmentProd,
		),
	)

	return func(cfg *Config) {
		cfg.Addr = *addr
		cfg.Environment = *env
	}
}

// loadServerFlags loads flags related to server settings
func loadServerFlags(flagSet *flag.FlagSet, cfg *Config) PostParseSetter {
	readTimeout := flagSet.Int64("rt", defaultReadTimeout, "Specify the read timeout for the server, in seconds")
	writeTimeout := flagSet.Int64("wt", defaultWriteTimeout, "Specify the write timeout for the server, in seconds")
	idleTimeout := flagSet.Int64("it", defaultIdleTimeout, "Specify the idle timeout for server connections, in seconds")
	DBDSN := flagSet.String("d", defaultDBDSN, "Specify the Data Source Name for connecting to the database")
	fileStoragePath := flagSet.String("f", defaultFileStoragePath, "Specify the filename where current metric values will be saved")
	restore := flagSet.Bool("r", defaultRestore, "Enable or disable the restoration of previously saved values from a file upon server startup")
	storeInterval := flagSet.Int64("i", defaultStoreInterval, "Set the interval for saving the current server metrics to disk, in seconds")
	maxOpenConns := flagSet.Int("mo", defaultMaxOpenConns, "Specify the maximum number of open database connections")
	maxIdleConns := flagSet.Int("mi", defaultMaxIdleConns, "Specify the maximum number of idle database connections")
	connMaxLifetime := flagSet.Int64("ml", defaultConnMaxLifetime, "Specify the maximum lifetime of a database connection, in seconds")
	initialDelay := flagSet.Int64("id", defaultInitialDelay, "Initial delay before the first retry attempt, in seconds")
	maxRetries := flagSet.Int("mr", defaultMaxRetries, "Maximum number of retries for the operation")
	delayIncrement := flagSet.Int64("di", defaultDelayIncrement, "Incremental delay between retries, in seconds")

	return func(cfg *Config) {
		cfg.ReadTimeout = time.Duration(*readTimeout) * time.Second
		cfg.WriteTimeout = time.Duration(*writeTimeout) * time.Second
		cfg.IdleTimeout = time.Duration(*idleTimeout) * time.Second
		cfg.DBDSN = *DBDSN
		cfg.FileStoragePath = *fileStoragePath
		cfg.Restore = *restore
		cfg.StoreInterval = time.Duration(*storeInterval) * time.Second
		cfg.MaxOpenConns = *maxOpenConns
		cfg.MaxIdleConns = *maxIdleConns
		cfg.ConnMaxLifetime = time.Duration(*connMaxLifetime) * time.Second
		cfg.InitialDelay = time.Duration(*initialDelay) * time.Second
		cfg.MaxRetries = *maxRetries
		cfg.DelayIncrement = time.Duration(*delayIncrement) * time.Second
	}
}

// loadAgentFlags loads flags related to agent settings
func loadAgentFlags(flagSet *flag.FlagSet, cfg *Config) PostParseSetter {
	reportInterval := flagSet.Int64("r", defaultReportInterval, "Set the interval for sending metrics to the server, in seconds")
	pollInterval := flagSet.Int64("p", defaultPollInterval, "Set the interval for polling metrics from the runtime package, in seconds")

	return func(cfg *Config) {
		cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
		cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	}
}

// parseServerConfig creates a new Config and populates it with server-related settings
func ParseServerConfig() (*Config, error) {
	cfg := &Config{}
	if err := loadAndParseFlags(cfg, loadGeneralFlags, loadServerFlags); err != nil {
		return nil, err
	}
	return cfg, loadFromEnv(cfg)
}

// parseAgentConfig creates a new Config and populates it with agent-related settings
func ParseAgentConfig() (*Config, error) {
	cfg := &Config{}
	if err := loadAndParseFlags(cfg, loadGeneralFlags, loadAgentFlags); err != nil {
		return nil, err
	}
	return cfg, loadFromEnv(cfg)
}

// loadFromEnv overrides Config fields from environment variables
func loadFromEnv(cfg *Config) error {
	if err := parseEnvWithDuration(cfg); err != nil {
		return err
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
		envName := field.Tag.Get("env")
		if envName == "" {
			continue
		}

		envValue := os.Getenv(envName)
		if envValue == "" {
			continue
		}

		if field.Type == reflect.TypeOf(time.Duration(0)) {
			envVars[envName] = envValue + "s"
		} else {
			envVars[envName] = envValue
		}
	}

	return envVars
}

func isValidEnvironment(env string) bool {
	return env == defaultEnvironmentDev || env == defaultEnvironmentProd
}
