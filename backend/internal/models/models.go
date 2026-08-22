package models

import "time"

// Bot statuses.
const (
	StatusActive = "active"
	StatusPaused = "paused"
)

// Built-in template identifiers. Every name here has a matching
// implementation registered by the engine.
const (
	TemplateAniTez          = "anitez"
	TemplateAniXUltra       = "anixultra"
	TemplateAIAssistant     = "ai_assistant"
	TemplateEcommerceShop   = "ecommerce_shop"
	TemplateFeedbackSupport = "feedback_support"
	TemplateChannelManager  = "channel_manager"
	TemplateCustomBuilder   = "custom_builder"
)

// AllTemplateNames returns every built-in template identifier.
func AllTemplateNames() []string {
	return []string{
		TemplateAniTez,
		TemplateAniXUltra,
		TemplateAIAssistant,
		TemplateEcommerceShop,
		TemplateFeedbackSupport,
		TemplateChannelManager,
		TemplateCustomBuilder,
	}
}

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

// BotUser is a Telegram user that interacted with a bot.
type BotUser struct {
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

// Platform account roles.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// User is a platform account that can sign in with Google or Telegram.
// Its balance is topped up through approved deposits and spent on
// platform services (e.g. custom repo templates).
type User struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id,omitempty"`
	GoogleID   string    `json:"google_id,omitempty"`
	Email      string    `json:"email,omitempty"`
	FullName   string    `json:"full_name,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Balance    int64     `json:"balance"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PaymentCard is a bank card shown to users as a manual transfer target.
type PaymentCard struct {
	ID         int64     `json:"id"`
	CardNumber string    `json:"card_number"`
	CardHolder string    `json:"card_holder"`
	BankName   string    `json:"bank_name"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// Deposit statuses.
const (
	DepositPending  = "pending"
	DepositApproved = "approved"
	DepositRejected = "rejected"
)

// Deposit is a balance top-up request backed by a payment receipt.
type Deposit struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	UserFullName string     `json:"user_full_name,omitempty"`
	Amount       int64      `json:"amount"`
	ReceiptURL   string     `json:"receipt_url,omitempty"`
	Status       string     `json:"status"`
	RejectReason string     `json:"reject_reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
}

// CustomRepoTemplate is a template submitted by an admin and backed by
// a git repository that the engine can clone.
type CustomRepoTemplate struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	GitRepoURL  string    `json:"git_repo_url"`
	Category    string    `json:"category,omitempty"`
	Price       int64     `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
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

// Order statuses.
const (
	OrderPending = "pending"
	OrderPaid    = "paid"
)

// OrderItem is one product line of an order.
type OrderItem struct {
	Code  string `json:"code"`
	Title string `json:"title"`
	Price int64  `json:"price"`
	Qty   int    `json:"qty"`
}

// Order is a purchase placed through an ecommerce bot. Items are stored
// as a JSON document in the database.
type Order struct {
	ID         int64       `json:"id"`
	BotID      int64       `json:"bot_id"`
	TelegramID int64       `json:"telegram_id"`
	Username   string      `json:"username,omitempty"`
	Items      []OrderItem `json:"items"`
	Total      int64       `json:"total"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
}

// Ticket statuses.
const (
	TicketOpen   = "open"
	TicketClosed = "closed"
)

// Ticket is a support conversation between a user and the bot admins.
type Ticket struct {
	ID         int64     `json:"id"`
	BotID      int64     `json:"bot_id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
