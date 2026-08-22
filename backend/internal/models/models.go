package models

import "time"

// Bot statuses.
const (
	StatusActive = "active"
	StatusPaused = "paused"
)

// Bot is a registered Telegram bot managed by the platform.
type Bot struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Username   string    `json:"username,omitempty"`
	Token      string    `json:"token"`
	Template   string    `json:"template"`
	WebhookURL string    `json:"webhook_url,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// IsActive reports whether the bot currently receives updates.
func (b *Bot) IsActive() bool {
	return b.Status == StatusActive
}

// Template describes a bot behaviour template exposed by the API.
type Template struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
}

// Stats aggregates counters for a single bot (or the whole platform).
type Stats struct {
	BotID            int64      `json:"bot_id,omitempty"`
	BotName          string     `json:"bot_name,omitempty"`
	TotalUsers       int        `json:"total_users"`
	NewUsersToday    int        `json:"new_users_today"`
	TotalMessages    int        `json:"total_messages"`
	TotalSearches    int        `json:"total_searches"`
	TotalCodeLookups int        `json:"total_code_lookups"`
	BroadcastsSent   int        `json:"broadcasts_sent"`
	LastActive       *time.Time `json:"last_active,omitempty"`
}

// User is a Telegram user that interacted with a bot.
type User struct {
	ID         int64     `json:"id"`
	BotID      int64     `json:"bot_id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username,omitempty"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	IsAdmin    bool      `json:"is_admin"`
	JoinedAt   time.Time `json:"joined_at"`
	LastSeen   time.Time `json:"last_seen"`
}

// Broadcast statuses.
const (
	BroadcastPending = "pending"
	BroadcastRunning = "running"
	BroadcastDone    = "done"
	BroadcastFailed  = "failed"
)

// Broadcast is a message being delivered to all users of a bot.
type Broadcast struct {
	ID        int64     `json:"id"`
	BotID     int64     `json:"bot_id"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Total     int       `json:"total"`
	Sent      int       `json:"sent"`
	Failed    int       `json:"failed"`
	CreatedAt time.Time `json:"created_at"`
}
