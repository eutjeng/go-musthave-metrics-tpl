package config

func isValidEnvironment(env string) bool {
	return env == defaultEnvironmentDev || env == defaultEnvironmentProd
}
