package mysocks

import (
	"os"
	"strconv"
)

func env(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

func intEnv(name string, defaultValue int) int {
	stringValue := env(name, "")
	if stringValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(stringValue)
	if err != nil {
		return defaultValue
	}
	return value
}

func portFromEnv() int {
	return intEnv("MYSOCKS_PORT", 1080)
}

func hostNameFromEnv() string {
	return env("MYSOCKS_HOSTNAME", "localhost")
}

func userNameFromEnv() string {
	return env("MYSOCKS_USER", "")
}

func passwordFromEnv() string {
	return env("MYSOCKS_PASSWORD", "")
}
