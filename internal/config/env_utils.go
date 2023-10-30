package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
)

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
