package utils

import "os"

// GetEnvOrDefault returns the environment variable or default string.
func GetEnvOrDefault(env string, defaultStr string) string {
	if env == "" {
		return defaultStr
	}

	fetchedEnv := os.Getenv(env)

	if fetchedEnv == "" {
		return defaultStr
	}

	return fetchedEnv
}
