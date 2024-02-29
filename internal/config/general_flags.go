package config

import (
	"flag"
	"fmt"
)

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
	key := flagSet.String("k", defaultKey, "Specify the secret key used for hashing. It should be a string value.")

	return func(cfg *Config) {
		cfg.Addr = *addr
		cfg.Environment = *env
		cfg.Key = *key
	}
}
