package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address        string
	DatabaseURL    string
	ArchiveRoot    string
	ShutdownWindow time.Duration
	Development    bool
}

func Load() (Config, error) {
	config := Config{Address: envOr("HTTP_ADDRESS", ":8080"), DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")), ArchiveRoot: envOr("ARCHIVE_ROOT", "./var/archives"), ShutdownWindow: 10 * time.Second}
	if raw := strings.TrimSpace(os.Getenv("SHUTDOWN_SECONDS")); raw != "" {
		seconds, err := strconv.Atoi(raw)
		if err != nil || seconds < 1 || seconds > 120 {
			return Config{}, fmt.Errorf("SHUTDOWN_SECONDS must be between 1 and 120")
		}
		config.ShutdownWindow = time.Duration(seconds) * time.Second
	}
	development, err := strconv.ParseBool(envOr("DEVELOPMENT", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("DEVELOPMENT must be a boolean: %w", err)
	}
	config.Development = development
	if !strings.Contains(config.Address, ":") {
		return Config{}, fmt.Errorf("HTTP_ADDRESS must include a port")
	}
	return config, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
