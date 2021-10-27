package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	ServerPort = getEnvWithDefault("SERVER_PORT", "3000")

	DBHost     = getEnvWithDefault("DATABASE_HOST", "localhost")
	DBPort     = getEnvWithDefault("DATABASE_PORT", "3306")
	DBName     = getEnvWithDefault("DATABASE_NAME", "url_shortener")
	DBUser     = getEnvWithDefault("DATABASE_USER", "root")
	DBPassword = getEnvWithDefault("DATABASE_PASSWORD", "root")

	RedisHost       = getEnvWithDefault("REDIS_HOST", "localhost")
	RedisPort       = getEnvWithDefault("REDIS_PORT", "6379")
	BloomFilterName = getEnvWithDefault("REDIS_BLOOM_FILTER_NAME", "url_shortener")

	CacheTTLInSeconds = getInt64EnvWithDefault("CACHE_TTL_IN_SECONDS", 86400)
)

func mustGetEnv(target string) string {
	if result, ok := os.LookupEnv(target); ok {
		return result
	}

	panic(fmt.Sprintf("Could not find environment variable: %s", target))
}

func getEnvWithDefault(target, def string) string {
	if result, ok := os.LookupEnv(target); ok {
		return result
	}

	return def
}

func getInt64EnvWithDefault(target string, def int64) int64 {
	if result, ok := os.LookupEnv(target); ok {
		res, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("invalid %s format", target))
		}

		return res
	}

	return def

}
