package config

import (
	"flag"
	"time"
)

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
