package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"botmaker-backend/internal/models"
)

// Store wraps the SQLite database used to persist bots, users,
// statistics and broadcasts.
type Store struct {
	db *sql.DB
}

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("storage: record not found")

// Open opens (creating if needed) the SQLite database at path and
// applies the schema.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite: single writer

	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure database: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS bots (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    username    TEXT    NOT NULL DEFAULT '',
    token       TEXT    NOT NULL UNIQUE,
    template    TEXT    NOT NULL,
    webhook_url TEXT    NOT NULL DEFAULT '',
    status      TEXT    NOT NULL DEFAULT 'active',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    bot_id      INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    telegram_id INTEGER NOT NULL,
    username    TEXT    NOT NULL DEFAULT '',
    first_name  TEXT    NOT NULL DEFAULT '',
    last_name   TEXT    NOT NULL DEFAULT '',
    is_admin    INTEGER NOT NULL DEFAULT 0,
    joined_at   INTEGER NOT NULL,
    last_seen   INTEGER NOT NULL,
    UNIQUE (bot_id, telegram_id)
);

CREATE TABLE IF NOT EXISTS stats (
    bot_id            INTEGER PRIMARY KEY REFERENCES bots(id) ON DELETE CASCADE,
    total_messages    INTEGER NOT NULL DEFAULT 0,
    total_searches    INTEGER NOT NULL DEFAULT 0,
    total_code_lookups INTEGER NOT NULL DEFAULT 0,
    last_active       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS broadcasts (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    bot_id     INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    message    TEXT    NOT NULL,
    status     TEXT    NOT NULL DEFAULT 'pending',
    total      INTEGER NOT NULL DEFAULT 0,
    sent       INTEGER NOT NULL DEFAULT 0,
    failed     INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS bot_settings (
    bot_id INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    key    TEXT    NOT NULL,
    value  TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY (bot_id, key)
);

CREATE TABLE IF NOT EXISTS orders (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    bot_id      INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    telegram_id INTEGER NOT NULL,
    username    TEXT    NOT NULL DEFAULT '',
    items       TEXT    NOT NULL DEFAULT '[]',
    total       INTEGER NOT NULL DEFAULT 0,
    status      TEXT    NOT NULL DEFAULT 'pending',
    created_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS tickets (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    bot_id      INTEGER NOT NULL REFERENCES bots(id) ON DELETE CASCADE,
    telegram_id INTEGER NOT NULL,
    username    TEXT    NOT NULL DEFAULT '',
    status      TEXT    NOT NULL DEFAULT 'open',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL,
    UNIQUE (bot_id, telegram_id)
);

CREATE INDEX IF NOT EXISTS idx_users_bot ON users(bot_id);
CREATE INDEX IF NOT EXISTS idx_broadcasts_bot ON broadcasts(bot_id);
CREATE INDEX IF NOT EXISTS idx_orders_bot ON orders(bot_id);
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(bot_id, telegram_id);
CREATE INDEX IF NOT EXISTS idx_tickets_bot ON tickets(bot_id);
`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

// ---- Bots ----

// CreateBot inserts a new bot and sets its ID.
func (s *Store) CreateBot(b *models.Bot) error {
	now := time.Now().Unix()
	res, err := s.db.Exec(
		`INSERT INTO bots (name, username, token, template, webhook_url, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.Name, b.Username, b.Token, b.Template, b.WebhookURL, b.Status, now, now,
	)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	b.ID = id
	b.CreatedAt = time.Unix(now, 0)
	b.UpdatedAt = time.Unix(now, 0)

	// Seed the stats row so counters can be incremented unconditionally.
	_, err = s.db.Exec(`INSERT OR IGNORE INTO stats (bot_id) VALUES (?)`, id)
	if err != nil {
		return fmt.Errorf("seed stats: %w", err)
	}
	return nil
}

// GetBot returns a bot by ID.
func (s *Store) GetBot(id int64) (*models.Bot, error) {
	row := s.db.QueryRow(
		`SELECT id, name, username, token, template, webhook_url, status, created_at, updated_at
		 FROM bots WHERE id = ?`, id)
	return scanBot(row)
}

// GetBotByToken returns a bot by its Telegram token.
func (s *Store) GetBotByToken(token string) (*models.Bot, error) {
	row := s.db.QueryRow(
		`SELECT id, name, username, token, template, webhook_url, status, created_at, updated_at
		 FROM bots WHERE token = ?`, token)
	return scanBot(row)
}

// ListBots returns all registered bots, newest first.
func (s *Store) ListBots() ([]models.Bot, error) {
	rows, err := s.db.Query(
		`SELECT id, name, username, token, template, webhook_url, status, created_at, updated_at
		 FROM bots ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list bots: %w", err)
	}
	defer rows.Close()

	var bots []models.Bot
	for rows.Next() {
		b, err := scanBot(rows)
		if err != nil {
			return nil, err
		}
		bots = append(bots, *b)
	}
	if bots == nil {
		bots = []models.Bot{}
	}
	return bots, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanBot(row scanner) (*models.Bot, error) {
	var (
		b         models.Bot
		createdAt int64
		updatedAt int64
	)
	err := row.Scan(&b.ID, &b.Name, &b.Username, &b.Token, &b.Template,
		&b.WebhookURL, &b.Status, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan bot: %w", err)
	}
	b.CreatedAt = time.Unix(createdAt, 0)
	b.UpdatedAt = time.Unix(updatedAt, 0)
	return &b, nil
}

// SetBotStatus updates the status of a bot.
func (s *Store) SetBotStatus(id int64, status string) error {
	_, err := s.db.Exec(
		`UPDATE bots SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("set bot status: %w", err)
	}
	return nil
}

// SetBotWebhook stores the webhook URL registered with Telegram.
func (s *Store) SetBotWebhook(id int64, url string) error {
	_, err := s.db.Exec(
		`UPDATE bots SET webhook_url = ?, updated_at = ? WHERE id = ?`,
		url, time.Now().Unix(), id)
	if err != nil {
		return fmt.Errorf("set bot webhook: %w", err)
	}
	return nil
}

// DeleteBot removes a bot and all of its dependent records.
func (s *Store) DeleteBot(id int64) error {
	_, err := s.db.Exec(`DELETE FROM bots WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete bot: %w", err)
	}
	return nil
}

// ---- Users ----

// UpsertUser records (or refreshes) a user interaction and reports
// whether the user is new to this bot.
func (s *Store) UpsertUser(u *models.User) (bool, error) {
	now := time.Now().Unix()
	isAdmin := 0
	if u.IsAdmin {
		isAdmin = 1
	}
	res, err := s.db.Exec(
		`INSERT INTO users (bot_id, telegram_id, username, first_name, last_name, is_admin, joined_at, last_seen)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (bot_id, telegram_id) DO UPDATE SET
		     username = excluded.username,
		     first_name = excluded.first_name,
		     last_name = excluded.last_name,
		     is_admin = excluded.is_admin,
		     last_seen = excluded.last_seen`,
		u.BotID, u.TelegramID, u.Username, u.FirstName, u.LastName, isAdmin, now, now,
	)
	if err != nil {
		return false, fmt.Errorf("upsert user: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	// A brand-new row reports 1 affected row, an update reports 2 in SQLite.
	return affected == 1, nil
}

// ListUserTelegramIDs returns the Telegram IDs of every user of a bot.
func (s *Store) ListUserTelegramIDs(botID int64) ([]int64, error) {
	rows, err := s.db.Query(`SELECT telegram_id FROM users WHERE bot_id = ?`, botID)
	if err != nil {
		return nil, fmt.Errorf("list user ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CountUsers returns the number of users of a bot.
func (s *Store) CountUsers(botID int64) (int, error) {
	return s.count(`SELECT COUNT(*) FROM users WHERE bot_id = ?`, botID)
}

// CountNewUsersToday returns how many users joined a bot since local midnight.
func (s *Store) CountNewUsersToday(botID int64) (int, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return s.count(
		`SELECT COUNT(*) FROM users WHERE bot_id = ? AND joined_at >= ?`,
		botID, start.Unix())
}

func (s *Store) count(query string, args ...any) (int, error) {
	var n int
	if err := s.db.QueryRow(query, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}
	return n, nil
}

// ---- Stats ----

// RecordMessage increments the total message counter of a bot.
func (s *Store) RecordMessage(botID int64) error {
	return s.bump(botID, "total_messages")
}

// RecordSearch increments the search counter of a bot.
func (s *Store) RecordSearch(botID int64) error {
	return s.bump(botID, "total_searches")
}

// RecordCodeLookup increments the code lookup counter of a bot.
func (s *Store) RecordCodeLookup(botID int64) error {
	return s.bump(botID, "total_code_lookups")
}

// TouchActivity updates the last activity timestamp of a bot.
func (s *Store) TouchActivity(botID int64) error {
	_, err := s.db.Exec(
		`UPDATE stats SET last_active = ? WHERE bot_id = ?`,
		time.Now().Unix(), botID)
	if err != nil {
		return fmt.Errorf("touch activity: %w", err)
	}
	return nil
}

func (s *Store) bump(botID int64, column string) error {
	_, err := s.db.Exec(
		fmt.Sprintf(`UPDATE stats SET %s = %s + 1, last_active = ? WHERE bot_id = ?`, column, column),
		time.Now().Unix(), botID)
	if err != nil {
		return fmt.Errorf("bump %s: %w", column, err)
	}
	return nil
}

// GetStats returns the aggregated statistics of a single bot.
func (s *Store) GetStats(botID int64) (*models.Stats, error) {
	b, err := s.GetBot(botID)
	if err != nil {
		return nil, err
	}
	return s.statsForBot(b)
}

func (s *Store) statsForBot(b *models.Bot) (*models.Stats, error) {
	var (
		totalUsers    int
		newUsersToday int
		messages      int
		searches      int
		lookups       int
		lastActive    int64
	)
	err := s.db.QueryRow(
		`SELECT COALESCE(total_messages,0), COALESCE(total_searches,0), COALESCE(total_code_lookups,0), COALESCE(last_active,0)
		 FROM stats WHERE bot_id = ?`, b.ID).Scan(&messages, &searches, &lookups, &lastActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	if totalUsers, err = s.CountUsers(b.ID); err != nil {
		return nil, err
	}
	if newUsersToday, err = s.CountNewUsersToday(b.ID); err != nil {
		return nil, err
	}

	broadcasts, err := s.count(
		`SELECT COUNT(*) FROM broadcasts WHERE bot_id = ? AND status = ?`, b.ID, models.BroadcastDone)
	if err != nil {
		return nil, err
	}

	st := &models.Stats{
		BotID:            b.ID,
		BotName:          b.Name,
		TotalUsers:       totalUsers,
		NewUsersToday:    newUsersToday,
		TotalMessages:    messages,
		TotalSearches:    searches,
		TotalCodeLookups: lookups,
		BroadcastsSent:   broadcasts,
	}
	if lastActive > 0 {
		t := time.Unix(lastActive, 0)
		st.LastActive = &t
	}
	return st, nil
}

// GetGlobalStats aggregates counters across every bot.
func (s *Store) GetGlobalStats() (*models.Stats, []models.Stats, error) {
	bots, err := s.ListBots()
	if err != nil {
		return nil, nil, err
	}

	global := &models.Stats{}
	perBot := []models.Stats{}
	for i := range bots {
		st, err := s.statsForBot(&bots[i])
		if err != nil {
			return nil, nil, err
		}
		perBot = append(perBot, *st)
		global.TotalUsers += st.TotalUsers
		global.NewUsersToday += st.NewUsersToday
		global.TotalMessages += st.TotalMessages
		global.TotalSearches += st.TotalSearches
		global.TotalCodeLookups += st.TotalCodeLookups
		global.BroadcastsSent += st.BroadcastsSent
	}
	return global, perBot, nil
}

// ---- Broadcasts ----

// CreateBroadcast inserts a broadcast and sets its ID.
func (s *Store) CreateBroadcast(b *models.Broadcast) error {
	res, err := s.db.Exec(
		`INSERT INTO broadcasts (bot_id, message, status, total, sent, failed, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.BotID, b.Message, b.Status, b.Total, b.Sent, b.Failed, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("create broadcast: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("create broadcast: %w", err)
	}
	b.ID = id
	return nil
}

// UpdateBroadcast persists the current progress of a broadcast.
func (s *Store) UpdateBroadcast(b *models.Broadcast) error {
	_, err := s.db.Exec(
		`UPDATE broadcasts SET status = ?, total = ?, sent = ?, failed = ? WHERE id = ?`,
		b.Status, b.Total, b.Sent, b.Failed, b.ID)
	if err != nil {
		return fmt.Errorf("update broadcast: %w", err)
	}
	return nil
}

// ListBroadcasts returns the broadcasts of a bot, newest first.
func (s *Store) ListBroadcasts(botID int64) ([]models.Broadcast, error) {
	rows, err := s.db.Query(
		`SELECT id, bot_id, message, status, total, sent, failed, created_at
		 FROM broadcasts WHERE bot_id = ? ORDER BY id DESC`, botID)
	if err != nil {
		return nil, fmt.Errorf("list broadcasts: %w", err)
	}
	defer rows.Close()

	var list []models.Broadcast
	for rows.Next() {
		var (
			b         models.Broadcast
			createdAt int64
		)
		if err := rows.Scan(&b.ID, &b.BotID, &b.Message, &b.Status,
			&b.Total, &b.Sent, &b.Failed, &createdAt); err != nil {
			return nil, err
		}
		b.CreatedAt = time.Unix(createdAt, 0)
		list = append(list, b)
	}
	if list == nil {
		list = []models.Broadcast{}
	}
	return list, rows.Err()
}

// ---- Bot settings ----

// GetSetting returns the value of a per-bot setting. The returned error
// wraps ErrNotFound when the setting has never been stored.
func (s *Store) GetSetting(botID int64, key string) (string, error) {
	var value string
	err := s.db.QueryRow(
		`SELECT value FROM bot_settings WHERE bot_id = ? AND key = ?`, botID, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get setting: %w", err)
	}
	return value, nil
}

// SetSetting upserts a per-bot setting.
func (s *Store) SetSetting(botID int64, key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO bot_settings (bot_id, key, value) VALUES (?, ?, ?)
		 ON CONFLICT (bot_id, key) DO UPDATE SET value = excluded.value`,
		botID, key, value)
	if err != nil {
		return fmt.Errorf("set setting: %w", err)
	}
	return nil
}

// DeleteSetting removes a per-bot setting.
func (s *Store) DeleteSetting(botID int64, key string) error {
	_, err := s.db.Exec(
		`DELETE FROM bot_settings WHERE bot_id = ? AND key = ?`, botID, key)
	if err != nil {
		return fmt.Errorf("delete setting: %w", err)
	}
	return nil
}

// ---- Orders ----

// CreateOrder inserts an order and sets its ID.
func (s *Store) CreateOrder(o *models.Order) error {
	items, err := json.Marshal(o.Items)
	if err != nil {
		return fmt.Errorf("encode order items: %w", err)
	}
	res, err := s.db.Exec(
		`INSERT INTO orders (bot_id, telegram_id, username, items, total, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		o.BotID, o.TelegramID, o.Username, string(items), o.Total, o.Status, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	o.ID = id
	return nil
}

// UpdateOrderStatus transitions an order to a new status.
func (s *Store) UpdateOrderStatus(id int64, status string) error {
	_, err := s.db.Exec(
		`UPDATE orders SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

// ListOrdersByUser returns the orders placed by a single user, newest first.
func (s *Store) ListOrdersByUser(botID, telegramID int64) ([]models.Order, error) {
	rows, err := s.db.Query(
		`SELECT id, bot_id, telegram_id, username, items, total, status, created_at
		 FROM orders WHERE bot_id = ? AND telegram_id = ? ORDER BY id DESC`, botID, telegramID)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()
	return scanOrders(rows)
}

func scanOrders(rows *sql.Rows) ([]models.Order, error) {
	var list []models.Order
	for rows.Next() {
		var (
			o         models.Order
			items     string
			createdAt int64
		)
		if err := rows.Scan(&o.ID, &o.BotID, &o.TelegramID, &o.Username,
			&items, &o.Total, &o.Status, &createdAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(items), &o.Items); err != nil {
			return nil, fmt.Errorf("decode order items: %w", err)
		}
		o.CreatedAt = time.Unix(createdAt, 0)
		list = append(list, o)
	}
	if list == nil {
		list = []models.Order{}
	}
	return list, rows.Err()
}

// ---- Tickets ----

// UpsertTicket creates or refreshes the support ticket of a user.
func (s *Store) UpsertTicket(t *models.Ticket) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO tickets (bot_id, telegram_id, username, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (bot_id, telegram_id) DO UPDATE SET
		     username = excluded.username,
		     status = 'open',
		     updated_at = excluded.updated_at`,
		t.BotID, t.TelegramID, t.Username, t.Status, now, now)
	if err != nil {
		return fmt.Errorf("upsert ticket: %w", err)
	}
	return nil
}

// GetTicket returns the ticket of a single user.
func (s *Store) GetTicket(botID, telegramID int64) (*models.Ticket, error) {
	var (
		t         models.Ticket
		createdAt int64
		updatedAt int64
	)
	err := s.db.QueryRow(
		`SELECT id, bot_id, telegram_id, username, status, created_at, updated_at
		 FROM tickets WHERE bot_id = ? AND telegram_id = ?`, botID, telegramID).
		Scan(&t.ID, &t.BotID, &t.TelegramID, &t.Username, &t.Status, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get ticket: %w", err)
	}
	t.CreatedAt = time.Unix(createdAt, 0)
	t.UpdatedAt = time.Unix(updatedAt, 0)
	return &t, nil
}

// ListTickets returns the tickets of a bot, newest first.
func (s *Store) ListTickets(botID int64) ([]models.Ticket, error) {
	rows, err := s.db.Query(
		`SELECT id, bot_id, telegram_id, username, status, created_at, updated_at
		 FROM tickets WHERE bot_id = ? ORDER BY id DESC`, botID)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var list []models.Ticket
	for rows.Next() {
		var (
			t         models.Ticket
			createdAt int64
			updatedAt int64
		)
		if err := rows.Scan(&t.ID, &t.BotID, &t.TelegramID, &t.Username,
			&t.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt = time.Unix(createdAt, 0)
		t.UpdatedAt = time.Unix(updatedAt, 0)
		list = append(list, t)
	}
	if list == nil {
		list = []models.Ticket{}
	}
	return list, rows.Err()
}

// SetTicketStatus transitions a user's ticket to a new status.
func (s *Store) SetTicketStatus(botID, telegramID int64, status string) error {
	_, err := s.db.Exec(
		`UPDATE tickets SET status = ?, updated_at = ? WHERE bot_id = ? AND telegram_id = ?`,
		status, time.Now().Unix(), botID, telegramID)
	if err != nil {
		return fmt.Errorf("set ticket status: %w", err)
	}
	return nil
}
