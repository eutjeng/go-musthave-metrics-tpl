package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type PostParseSetter func(*Config)
type FlagSetter func(*flag.FlagSet, *Config) PostParseSetter

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
	Key             string        `env:"KEY"`               // the secret key used for hashing data before transmission
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS"`    // max number of open database connections
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS"`    // max number of idle database connections
	RateLimit       int           `env:"RATE_LIMIT"`
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
