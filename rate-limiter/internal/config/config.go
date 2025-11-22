package config

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RateLimitIP    int64
	RateLimitToken int64
	BlockTime      int64
	ServerPort     string
}

func LoadConfig() *Config {
	return &Config{
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvAsInt("REDIS_DB", 0),
		RateLimitIP:    int64(getEnvAsInt("RATE_LIMIT_IP", 5)),
		RateLimitToken: int64(getEnvAsInt("RATE_LIMIT_TOKEN", 10)),
		BlockTime:      int64(getEnvAsInt("BLOCK_TIME", 300)),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
