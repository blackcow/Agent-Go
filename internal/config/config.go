package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 描述服务运行时配置。
type Config struct {
	Port             string
	CacheTTL         time.Duration
	RequestTimeout   time.Duration
	HistorySize      int
	DefaultProviders []string
}

// Load 从环境变量读取配置。
func Load() Config {
	cfg := Config{
		Port:             getEnv("APP_PORT", "8080"),
		CacheTTL:         parseDuration("CACHE_TTL", 5*time.Minute),
		RequestTimeout:   parseDuration("REQUEST_TIMEOUT", 8*time.Second),
		HistorySize:      parseInt("HISTORY_SIZE", 50),
		DefaultProviders: parseList("PROVIDERS", []string{"mock"}),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}

func parseInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func parseList(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
