package env

import "os"

// GetOrDefault returns the environment variable or default string.
func GetOrDefault(env string, defaultStr string) string {
	if env == "" {
		return defaultStr
	}

	fetchedEnv := os.Getenv(env)

	if fetchedEnv == "" {
		return defaultStr
	}

	return fetchedEnv
}
