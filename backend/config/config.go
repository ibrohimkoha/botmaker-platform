package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the botmaker platform.
type Config struct {
	// Port is the HTTP listen port of the management API.
	Port string
	// DBPath is the path to the SQLite database file.
	DBPath string
	// WebhookBaseURL is the public base URL that Telegram will call,
	// e.g. "https://bot.example.com". Each bot gets its own path:
	// <WebhookBaseURL>/api/webhook/<token>.
	WebhookBaseURL string
	// AdminIDs are Telegram user IDs allowed to use admin bot commands.
	AdminIDs []int64
}

// Load reads configuration from environment variables.
func Load() Config {
	cfg := Config{
		Port:           getenv("PORT", "8085"),
		DBPath:         getenv("DATABASE_PATH", "botmaker.db"),
		WebhookBaseURL: strings.TrimSuffix(getenv("WEBHOOK_BASE_URL", "http://localhost:8085"), "/"),
	}
	for _, part := range strings.Split(getenv("ADMIN_IDS", ""), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil {
			cfg.AdminIDs = append(cfg.AdminIDs, id)
		}
	}
	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
