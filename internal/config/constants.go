package config

const (
	defaultAddr            = ":8080"
	defaultEnvironmentDev  = "development"
	defaultEnvironmentProd = "production"
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultDBDSN           = ""
	defaultKey             = "supersecretkey"
	defaultRestore         = true
	defaultRateLimit       = 2
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
