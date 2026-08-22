package templates

import (
	"testing"

	"botmaker-backend/internal/models"
)

func TestCartLifecycle(t *testing.T) {
	cs := &cartState{carts: make(map[int64]map[int64]map[string]int)}

	const (
		botID = 1
		user  = 100
	)

	// Empty cart has no items.
	items, total := cs.snapshot(botID, user)
	if len(items) != 0 || total != 0 {
		t.Fatalf("expected empty cart, got %d items, %d total", len(items), total)
	}

	// Add two products.
	if err := cs.add(botID, user, "eco-001"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := cs.add(botID, user, "ECO-001"); err != nil {
		t.Fatalf("add (case-insensitive): %v", err)
	}
	if err := cs.add(botID, user, "ECO-012"); err != nil {
		t.Fatalf("add: %v", err)
	}

	items, total = cs.snapshot(botID, user)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Qty != 2 || items[1].Qty != 1 {
		t.Fatalf("unexpected quantities: %+v", items)
	}
	wantTotal := 2*12_900_000 + 120_000
	if total != int64(wantTotal) {
		t.Fatalf("total mismatch: got %d want %d", total, wantTotal)
	}

	// Unknown product codes are rejected.
	if err := cs.add(botID, user, "NOPE-001"); err == nil {
		t.Fatal("expected error for unknown product")
	}

	// Stock limit is enforced (ECO-001 has InStock 10).
	for i := 0; i < 12; i++ {
		_ = cs.add(botID, user, "ECO-001")
	}
	if err := cs.add(botID, user, "ECO-001"); err != errOutOfStock {
		t.Fatalf("expected errOutOfStock, got %v", err)
	}

	// Decrement removes the line when it reaches zero.
	cs.setQty(botID, user, "ECO-012", 0)
	items, _ = cs.snapshot(botID, user)
	if len(items) != 1 {
		t.Fatalf("expected 1 item after removal, got %d", len(items))
	}

	// Carts are per-user.
	cs.add(botID, user+1, "ECO-002")
	otherItems, _ := cs.snapshot(botID, user+1)
	if len(otherItems) != 1 || otherItems[0].Code != "ECO-002" {
		t.Fatalf("user carts mixed up: %+v", otherItems)
	}

	// Clear wipes only this user's cart.
	cs.clear(botID, user)
	items, _ = cs.snapshot(botID, user)
	if len(items) != 0 {
		t.Fatalf("expected empty cart after clear, got %d", len(items))
	}
	otherItems, _ = cs.snapshot(botID, user+1)
	if len(otherItems) != 1 {
		t.Fatalf("clear leaked into other user's cart: %+v", otherItems)
	}
}

func TestMoney(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1 000"},
		{12_900_000, "12 900 000"},
		{1_000_000_000, "1 000 000 000"},
	}
	for _, c := range cases {
		if got := money(c.in); got != c.want {
			t.Errorf("money(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestOrderItemsRoundTrip(t *testing.T) {
	order := &models.Order{
		Items: []models.OrderItem{
			{Code: "ECO-001", Title: "iPhone 15 Pro", Price: 12_900_000, Qty: 2},
		},
		Total:  25_800_000,
		Status: models.OrderPending,
	}
	if order.Status != models.OrderPending {
		t.Fatalf("unexpected status: %q", order.Status)
	}
	if len(order.Items) != 1 || order.Items[0].Code != "ECO-001" {
		t.Fatalf("unexpected items: %+v", order.Items)
	}
}
