package storage

import (
	"testing"

	"botmaker-backend/internal/models"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func createTestBot(t *testing.T, store *Store, name, template string) *models.Bot {
	t.Helper()
	bot := &models.Bot{
		Name:     name,
		Username: "test_" + name,
		Token:    "tok:" + name,
		Template: template,
		Status:   models.StatusActive,
	}
	if err := store.CreateBot(bot); err != nil {
		t.Fatalf("create bot: %v", err)
	}
	return bot
}

func TestBotSettings(t *testing.T) {
	store := openTestStore(t)
	bot := createTestBot(t, store, "chan", models.TemplateChannelManager)

	if _, err := store.GetSetting(bot.ID, "channels"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for missing setting, got %v", err)
	}

	want := `["@channel1","@channel2"]`
	if err := store.SetSetting(bot.ID, "channels", want); err != nil {
		t.Fatalf("set setting: %v", err)
	}
	got, err := store.GetSetting(bot.ID, "channels")
	if err != nil {
		t.Fatalf("get setting: %v", err)
	}
	if got != want {
		t.Fatalf("setting mismatch: got %q want %q", got, want)
	}

	// A second bot must not see the first bot's setting.
	other := createTestBot(t, store, "other", models.TemplateCustomBuilder)
	if _, err := store.GetSetting(other.ID, "channels"); err != ErrNotFound {
		t.Fatalf("setting leaked across bots: %v", err)
	}

	if err := store.DeleteSetting(bot.ID, "channels"); err != nil {
		t.Fatalf("delete setting: %v", err)
	}
	if _, err := store.GetSetting(bot.ID, "channels"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestOrders(t *testing.T) {
	store := openTestStore(t)
	bot := createTestBot(t, store, "shop", models.TemplateEcommerceShop)

	order := &models.Order{
		BotID:      bot.ID,
		TelegramID: 555,
		Username:   "alice",
		Items: []models.OrderItem{
			{Code: "ECO-001", Title: "iPhone 15 Pro", Price: 12_900_000, Qty: 2},
			{Code: "ECO-012", Title: "Kitob", Price: 120_000, Qty: 1},
		},
		Total:  25_920_000,
		Status: models.OrderPending,
	}
	if err := store.CreateOrder(order); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.ID == 0 {
		t.Fatal("create order did not assign an ID")
	}

	orders, err := store.ListOrdersByUser(bot.ID, 555)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	got := orders[0]
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
	if got.Total != 25_920_000 || got.Items[0].Qty != 2 {
		t.Fatalf("order fields wrong: total=%d qty=%d", got.Total, got.Items[0].Qty)
	}

	// Orders are scoped per user.
	others, err := store.ListOrdersByUser(bot.ID, 666)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(others) != 0 {
		t.Fatalf("expected no orders for other user, got %d", len(others))
	}

	if err := store.UpdateOrderStatus(order.ID, models.OrderPaid); err != nil {
		t.Fatalf("update order status: %v", err)
	}
	orders, _ = store.ListOrdersByUser(bot.ID, 555)
	if orders[0].Status != models.OrderPaid {
		t.Fatalf("expected paid status, got %q", orders[0].Status)
	}
}

func TestTickets(t *testing.T) {
	store := openTestStore(t)
	bot := createTestBot(t, store, "support", models.TemplateFeedbackSupport)

	tk := &models.Ticket{
		BotID:      bot.ID,
		TelegramID: 777,
		Username:   "bob",
		Status:     models.TicketOpen,
	}
	if err := store.UpsertTicket(tk); err != nil {
		t.Fatalf("upsert ticket: %v", err)
	}

	got, err := store.GetTicket(bot.ID, 777)
	if err != nil {
		t.Fatalf("get ticket: %v", err)
	}
	if got.Status != models.TicketOpen || got.Username != "bob" {
		t.Fatalf("ticket fields wrong: %+v", got)
	}

	// Re-messaging refreshes the ticket back to open and keeps it single.
	if err := store.SetTicketStatus(bot.ID, 777, models.TicketClosed); err != nil {
		t.Fatalf("set ticket status: %v", err)
	}
	got, _ = store.GetTicket(bot.ID, 777)
	if got.Status != models.TicketClosed {
		t.Fatalf("expected closed status, got %q", got.Status)
	}

	tk.Username = "bob2"
	if err := store.UpsertTicket(tk); err != nil {
		t.Fatalf("upsert ticket: %v", err)
	}
	got, _ = store.GetTicket(bot.ID, 777)
	if got.Status != models.TicketOpen || got.Username != "bob2" {
		t.Fatalf("ticket not refreshed: %+v", got)
	}

	tickets, err := store.ListTickets(bot.ID)
	if err != nil {
		t.Fatalf("list tickets: %v", err)
	}
	if len(tickets) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(tickets))
	}

	// Deleting the bot cascades to its tickets.
	if err := store.DeleteBot(bot.ID); err != nil {
		t.Fatalf("delete bot: %v", err)
	}
	tickets, _ = store.ListTickets(bot.ID)
	if len(tickets) != 0 {
		t.Fatalf("expected no tickets after bot delete, got %d", len(tickets))
	}
}
