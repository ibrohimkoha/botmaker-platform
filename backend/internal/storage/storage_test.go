package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"botmaker-backend/internal/models"
)

// TestMigrateLegacyUsers verifies that a database created with the old
// per-bot "users" table is migrated to "bot_users" without data loss,
// and that the new platform "users" table gets the new schema.
func TestMigrateLegacyUsers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE bots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL, username TEXT NOT NULL DEFAULT '',
			token TEXT NOT NULL UNIQUE, template TEXT NOT NULL,
			webhook_url TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'active',
			created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
		);
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bot_id INTEGER NOT NULL, telegram_id INTEGER NOT NULL,
			username TEXT NOT NULL DEFAULT '', first_name TEXT NOT NULL DEFAULT '',
			last_name TEXT NOT NULL DEFAULT '', is_admin INTEGER NOT NULL DEFAULT 0,
			joined_at INTEGER NOT NULL, last_seen INTEGER NOT NULL
		);
		INSERT INTO bots (name, token, template, created_at, updated_at)
			VALUES ('test', 'token-legacy', 'anitez', 1, 1);
		INSERT INTO users (bot_id, telegram_id, username, joined_at, last_seen)
			VALUES (1, 42, 'alice', 1, 1);
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	store, err := Open(path)
	if err != nil {
		t.Fatalf("open after migration: %v", err)
	}
	defer store.Close()

	// The legacy row must survive in bot_users.
	var tgID int64
	if err := store.db.QueryRow(
		`SELECT telegram_id FROM bot_users WHERE bot_id = 1`).Scan(&tgID); err != nil {
		t.Fatalf("legacy row lost: %v", err)
	}
	if tgID != 42 {
		t.Fatalf("legacy telegram_id = %d, want 42", tgID)
	}

	// The platform users table must be empty and writable with the new schema.
	if _, err := store.GetUserByTelegramID(42); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetUserByTelegramID(42) = %v, want ErrNotFound", err)
	}
	u, err := store.UpsertTelegramUser(&models.User{
		TelegramID: 5415350162,
		FullName:   "Super Admin",
		Role:       models.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("upsert telegram user: %v", err)
	}
	if u.ID == 0 || u.Role != models.RoleAdmin || u.Balance != 0 {
		t.Fatalf("unexpected user after upsert: %+v", u)
	}
}

// TestNewSchemaOperations exercises the platform accounts, payment cards,
// deposits and custom template storage on a fresh database.
func TestNewSchemaOperations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fresh.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Accounts and balances.
	u, err := store.UpsertGoogleUser(&models.User{
		GoogleID:  "google-sub-1",
		Email:     "user@example.com",
		FullName:  "Test User",
		AvatarURL: "https://example.com/avatar.png",
		Role:      models.RoleUser,
	})
	if err != nil {
		t.Fatalf("upsert google user: %v", err)
	}
	if err := store.AddBalance(u.ID, 50000); err != nil {
		t.Fatalf("add balance: %v", err)
	}
	got, err := store.GetUser(u.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if got.Balance != 50000 {
		t.Fatalf("balance = %d, want 50000", got.Balance)
	}
	if err := store.DeductBalance(u.ID, 30000); err != nil {
		t.Fatalf("deduct balance: %v", err)
	}
	got, _ = store.GetUser(u.ID)
	if got.Balance != 20000 {
		t.Fatalf("balance after deduct = %d, want 20000", got.Balance)
	}
	if err := store.DeductBalance(u.ID, 999999); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("deduct too much = %v, want ErrInsufficientBalance", err)
	}

	// Payment cards.
	card := &models.PaymentCard{CardNumber: "8600 0000 0000 0000", CardHolder: "Ivan", BankName: "Humo", IsActive: true}
	if err := store.CreateCard(card); err != nil {
		t.Fatalf("create card: %v", err)
	}
	active, err := store.ListActiveCards()
	if err != nil || len(active) != 1 {
		t.Fatalf("list active cards = %v, %v; want 1 card", active, err)
	}
	if err := store.SetCardActive(card.ID, false); err != nil {
		t.Fatalf("set card inactive: %v", err)
	}
	if cards, _ := store.ListActiveCards(); len(cards) != 0 {
		t.Fatalf("expected no active cards, got %d", len(cards))
	}

	// Deposits: approval credits the balance.
	d := &models.Deposit{UserID: u.ID, UserFullName: u.FullName, Amount: 100000, ReceiptURL: "https://b2.example.com/receipt.png", Status: models.DepositPending}
	if err := store.CreateDeposit(d); err != nil {
		t.Fatalf("create deposit: %v", err)
	}
	if err := store.ApproveDeposit(d.ID); err != nil {
		t.Fatalf("approve deposit: %v", err)
	}
	got, _ = store.GetUser(u.ID)
	if got.Balance != 120000 {
		t.Fatalf("balance after approval = %d, want 120000", got.Balance)
	}
	dep, err := store.GetDeposit(d.ID)
	if err != nil || dep.Status != models.DepositApproved || dep.ApprovedAt == nil {
		t.Fatalf("deposit after approval = %+v, %v", dep, err)
	}
	if err := store.ApproveDeposit(d.ID); !errors.Is(err, ErrDepositProcessed) {
		t.Fatalf("double approve = %v, want ErrDepositProcessed", err)
	}
	d2 := &models.Deposit{UserID: u.ID, Amount: 5000, Status: models.DepositPending}
	if err := store.CreateDeposit(d2); err != nil {
		t.Fatal(err)
	}
	if err := store.RejectDeposit(d2.ID, "chek aniq emas"); err != nil {
		t.Fatalf("reject deposit: %v", err)
	}
	dep2, _ := store.GetDeposit(d2.ID)
	if dep2.Status != models.DepositRejected || dep2.RejectReason != "chek aniq emas" {
		t.Fatalf("deposit after rejection = %+v", dep2)
	}
	// Balance unchanged by rejection.
	got, _ = store.GetUser(u.ID)
	if got.Balance != 120000 {
		t.Fatalf("balance after rejection = %d, want 120000", got.Balance)
	}
	// Per-user listing.
	mine, err := store.ListDepositsByUser(u.ID)
	if err != nil || len(mine) != 2 {
		t.Fatalf("list deposits by user = %v, %v; want 2", mine, err)
	}

	// Custom repo templates.
	tpl := &models.CustomRepoTemplate{Name: "my-tpl", Title: "My Template", GitRepoURL: "https://github.com/x/y", Category: "biznes", Price: 25000}
	if err := store.CreateCustomRepoTemplate(tpl); err != nil {
		t.Fatalf("create template: %v", err)
	}
	tpls, err := store.ListCustomRepoTemplates()
	if err != nil || len(tpls) != 1 || tpls[0].Price != 25000 {
		t.Fatalf("list templates = %v, %v; want 1", tpls, err)
	}

	// Bot settings round-trip.
	bot := &models.Bot{Name: "settings bot", Token: "token-settings", Template: "anitez", Status: models.StatusActive}
	if err := store.CreateBot(bot); err != nil {
		t.Fatalf("create bot: %v", err)
	}
	if err := store.SetAllSettings(bot.ID, map[string]string{"admin_id": "12", "currency": "UZS"}); err != nil {
		t.Fatalf("set all settings: %v", err)
	}
	settings, err := store.GetAllSettings(bot.ID)
	if err != nil || settings["admin_id"] != "12" || settings["currency"] != "UZS" {
		t.Fatalf("get all settings = %v, %v", settings, err)
	}
}
