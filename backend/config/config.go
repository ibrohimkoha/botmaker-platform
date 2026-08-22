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
	// AI holds the LLM settings used by the ai_assistant template.
	AI AIConfig
}

// AIConfig configures the OpenAI-compatible chat API (DeepSeek by default)
// used by the ai_assistant template.
type AIConfig struct {
	// BaseURL is the API root, e.g. "https://api.deepseek.com".
	// The chat completions endpoint is appended to it.
	BaseURL string
	// APIKey is the bearer token sent to the AI provider.
	APIKey string
	// Model is the model identifier, e.g. "deepseek-chat".
	Model string
}

// Load reads configuration from environment variables.
func Load() Config {
	cfg := Config{
		Port:           getenv("PORT", "8085"),
		DBPath:         getenv("DATABASE_PATH", "botmaker.db"),
		WebhookBaseURL: strings.TrimSuffix(getenv("WEBHOOK_BASE_URL", "http://localhost:8085"), "/"),
		AI: AIConfig{
			BaseURL: strings.TrimSuffix(getenv("AI_BASE_URL", "https://api.deepseek.com"), "/"),
			APIKey:  getenv("AI_API_KEY", ""),
			Model:   getenv("AI_MODEL", "deepseek-chat"),
		},
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
