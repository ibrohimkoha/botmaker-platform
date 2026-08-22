package templates

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/internal/models"
)

// EcommerceShop is the online store template: product catalog, cart,
// order placement and simulated payment.
type EcommerceShop struct {
	carts *cartState
}

// Product is one item of the ecommerce catalog.
type Product struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Price       int64  `json:"price"` // in so'm
	Description string `json:"description"`
	InStock     int    `json:"in_stock,omitempty"` // 0 = unlimited
}

// cartState holds the per-user carts of a bot.
type cartState struct {
	mu    sync.Mutex
	carts map[int64]map[int64]map[string]int // botID -> userID -> code -> qty
}

func (cs *cartState) add(botID, userID int64, code string) error {
	p, ok := findProduct(code)
	if !ok {
		return errProductNotFound
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.carts[botID] == nil {
		cs.carts[botID] = make(map[int64]map[string]int)
	}
	if cs.carts[botID][userID] == nil {
		cs.carts[botID][userID] = make(map[string]int)
	}
	cur := cs.carts[botID][userID][p.Code]
	if p.InStock > 0 && cur >= p.InStock {
		return errOutOfStock
	}
	cs.carts[botID][userID][p.Code] = cur + 1
	return nil
}

func (cs *cartState) qty(botID, userID int64, code string) int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.carts[botID][userID][code]
}

func (cs *cartState) setQty(botID, userID int64, code string, qty int) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.carts[botID] == nil || cs.carts[botID][userID] == nil {
		return
	}
	if qty <= 0 {
		delete(cs.carts[botID][userID], code)
		return
	}
	cs.carts[botID][userID][code] = qty
}

func (cs *cartState) count(botID, userID int64) int {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	n := 0
	for _, qty := range cs.carts[botID][userID] {
		n += qty
	}
	return n
}

func (cs *cartState) clear(botID, userID int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.carts[botID], userID)
}

// snapshot returns the current cart as order items plus the total price.
func (cs *cartState) snapshot(botID, userID int64) ([]models.OrderItem, int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	codes := make([]string, 0, len(cs.carts[botID][userID]))
	for code := range cs.carts[botID][userID] {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	var (
		items []models.OrderItem
		total int64
	)
	for _, code := range codes {
		qty := cs.carts[botID][userID][code]
		p, ok := findProduct(code)
		if !ok {
			continue
		}
		items = append(items, models.OrderItem{
			Code: p.Code, Title: p.Title, Price: p.Price, Qty: qty,
		})
		total += p.Price * int64(qty)
	}
	return items, total
}

var (
	errProductNotFound = fmt.Errorf("product not found")
	errOutOfStock      = fmt.Errorf("out of stock")
	errEmptyCart       = fmt.Errorf("cart is empty")
	errOrderNotFound   = fmt.Errorf("order not found")
)

func findProduct(code string) (Product, bool) {
	code = normalizeCode(code)
	for _, p := range ecommerceCatalog {
		if normalizeCode(p.Code) == code {
			return p, true
		}
	}
	return Product{}, false
}

// Name returns the stable template identifier.
func (t *EcommerceShop) Name() string { return "ecommerce_shop" }

// Title returns the display name of the template.
func (t *EcommerceShop) Title() string { return "Ecommerce Shop" }

// Description returns the human readable summary used by the API.
func (t *EcommerceShop) Description() string {
	return "Online do'kon boti — katalog, savatcha, buyurtmalar va to'lov"
}

// Commands lists the commands exposed by this template.
func (t *EcommerceShop) Commands() []string {
	return []string{
		"/start", "/help",
		"/menu",
		"/cart",
		"/checkout",
		"/orders",
		"/pay <id>",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the EcommerceShop template on the bot.
func (t *EcommerceShop) Apply(bot *tele.Bot, opts Options) error {
	if t.carts == nil {
		t.carts = &cartState{carts: make(map[int64]map[int64]map[string]int)}
	}

	applyCommon(bot, "Ecommerce Shop", opts)

	bot.Handle("/start", func(c tele.Context) error {
		name := ""
		if sender := c.Sender(); sender != nil {
			name = sender.FirstName
			if name == "" {
				name = sender.Username
			}
		}
		return c.Send(fmt.Sprintf("👋 Salom, <b>%s</b>!\n\n🛍 <b>Ecommerce Shop</b> — onlayn do'kon boti.\n\nQuyidagi menyudan mahsulot tanlang yoki <code>/menu</code> buyrug'ini bosing.", esc(name)),
			catalogMarkup(t.carts.count(opts.BotID, senderID(c))))
	})

	bot.Handle("/help", func(c tele.Context) error {
		return c.Send(`🛍 <b>Ecommerce Shop — yordam</b>

• <code>/menu</code> — do'kon katalogi
• <code>/cart</code> — savatchani ko'rish
• <code>/checkout</code> — buyurtma rasmiylashtirish
• <code>/orders</code> — buyurtmalarim
• <code>/pay &lt;id&gt;</code> — buyurtma to'lovini tasdiqlash

🛡️ Adminlar uchun: /admin, /broadcast <matn>`)
	})

	bot.Handle("/menu", func(c tele.Context) error {
		return c.Send(catalogText(), catalogMarkup(t.carts.count(opts.BotID, senderID(c))))
	})

	bot.Handle("/cart", func(c tele.Context) error {
		return sendCart(c, t, opts)
	})

	bot.Handle("/checkout", func(c tele.Context) error {
		order, err := t.checkout(bot, opts, c.Sender())
		if err != nil {
			if err == errEmptyCart {
				return c.Send("🛒 Savatcha bo'sh. Avval mahsulot qo'shing: /menu")
			}
			return c.Send("❌ Buyurtma yaratib bo'lmadi.")
		}
		return c.Send(orderText(order), payMarkup(order))
	})

	bot.Handle("/orders", func(c tele.Context) error {
		orders, err := opts.Store.ListOrdersByUser(opts.BotID, senderID(c))
		if err != nil {
			return c.Send("❌ Buyurtmalarni o'qib bo'lmadi.")
		}
		if len(orders) == 0 {
			return c.Send("📦 Hozircha buyurtmalaringiz yo'q. Xarid qilish: /menu")
		}
		return c.Send(orderListText(orders))
	})

	bot.Handle("/pay", func(c tele.Context) error {
		id, err := strconv.ParseInt(strings.TrimSpace(c.Data()), 10, 64)
		if err != nil {
			return c.Send("ℹ️ Buyurtma id sini kiriting: <code>/pay 12</code>")
		}
		order, err := t.markPaid(opts, senderID(c), id)
		if err != nil {
			if err == errOrderNotFound {
				return c.Send("😕 Buyurtma topilmadi.")
			}
			return c.Send("❌ To'lovni tasdiqlab bo'lmadi.")
		}
		notifyAdmins(bot, opts, fmt.Sprintf("💳 <b>To'lov tasdiqlandi</b> — buyurtma #%d, %s", order.ID, money(order.Total)+" so'm"))
		return c.Send(orderPaidText(order), menuOnlyMarkup())
	})

	bot.Handle(tele.OnCallback, func(c tele.Context) error {
		sender := c.Sender()
		if sender == nil {
			return nil
		}
		uid := sender.ID
		data := c.Data()

		switch {
		case data == "menu":
			_ = c.Respond()
			return c.Edit(catalogText(), catalogMarkup(t.carts.count(opts.BotID, uid)))
		case data == "cart":
			_ = c.Respond()
			return c.Edit(cartText(t.carts, opts.BotID, uid), cartMarkup(t.carts, opts.BotID, uid))
		case data == "clear":
			t.carts.clear(opts.BotID, uid)
			_ = c.Respond(&tele.CallbackResponse{Text: "Savatcha tozalandi"})
			return c.Edit(cartText(t.carts, opts.BotID, uid), cartMarkup(t.carts, opts.BotID, uid))
		case data == "checkout":
			order, err := t.checkout(bot, opts, sender)
			if err != nil {
				if err == errEmptyCart {
					_ = c.Respond(&tele.CallbackResponse{Text: "Savatcha bo'sh", ShowAlert: true})
					return nil
				}
				_ = c.Respond(&tele.CallbackResponse{Text: "Xatolik yuz berdi", ShowAlert: true})
				return nil
			}
			_ = c.Respond(&tele.CallbackResponse{Text: "✅ Buyurtma qabul qilindi"})
			return c.Edit(orderText(order), payMarkup(order))
		case strings.HasPrefix(data, "add_"):
			code := strings.TrimPrefix(data, "add_")
			p, ok := findProduct(code)
			if !ok {
				_ = c.Respond(&tele.CallbackResponse{Text: "Mahsulot topilmadi", ShowAlert: true})
				return nil
			}
			if err := t.carts.add(opts.BotID, uid, p.Code); err != nil {
				if err == errOutOfStock {
					_ = c.Respond(&tele.CallbackResponse{Text: "Zaxirada yetarli emas", ShowAlert: true})
					return nil
				}
				_ = c.Respond()
				return nil
			}
			_ = c.Respond(&tele.CallbackResponse{Text: "Qo'shildi: " + p.Title})
			return nil
		case strings.HasPrefix(data, "inc_"):
			code := strings.TrimPrefix(data, "inc_")
			if err := t.carts.add(opts.BotID, uid, code); err == errOutOfStock {
				_ = c.Respond(&tele.CallbackResponse{Text: "Zaxirada yetarli emas", ShowAlert: true})
				return nil
			}
			_ = c.Respond()
			return c.Edit(cartText(t.carts, opts.BotID, uid), cartMarkup(t.carts, opts.BotID, uid))
		case strings.HasPrefix(data, "dec_"):
			code := strings.TrimPrefix(data, "dec_")
			t.carts.setQty(opts.BotID, uid, code, t.carts.qty(opts.BotID, uid, code)-1)
			_ = c.Respond()
			return c.Edit(cartText(t.carts, opts.BotID, uid), cartMarkup(t.carts, opts.BotID, uid))
		case strings.HasPrefix(data, "rem_"):
			t.carts.setQty(opts.BotID, uid, strings.TrimPrefix(data, "rem_"), 0)
			_ = c.Respond()
			return c.Edit(cartText(t.carts, opts.BotID, uid), cartMarkup(t.carts, opts.BotID, uid))
		case strings.HasPrefix(data, "pay_"):
			id, err := strconv.ParseInt(strings.TrimPrefix(data, "pay_"), 10, 64)
			if err != nil {
				_ = c.Respond()
				return nil
			}
			order, err := t.markPaid(opts, uid, id)
			if err != nil {
				_ = c.Respond(&tele.CallbackResponse{Text: "Buyurtma topilmadi", ShowAlert: true})
				return nil
			}
			notifyAdmins(bot, opts, fmt.Sprintf("💳 <b>To'lov tasdiqlandi</b> — buyurtma #%d, %s", order.ID, money(order.Total)+" so'm"))
			_ = c.Respond(&tele.CallbackResponse{Text: "✅ To'lov tasdiqlandi"})
			return c.Edit(orderPaidText(order), menuOnlyMarkup())
		default:
			_ = c.Respond()
			return nil
		}
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("ℹ️ Do'konni ochish uchun <code>/menu</code> ni bosing.")
	})
	return nil
}

func senderID(c tele.Context) int64 {
	if s := c.Sender(); s != nil {
		return s.ID
	}
	return 0
}

func sendCart(c tele.Context, t *EcommerceShop, opts Options) error {
	if t.carts.count(opts.BotID, senderID(c)) == 0 {
		return c.Send("🛒 Savatcha bo'sh. Mahsulot qo'shish: /menu")
	}
	return c.Send(cartText(t.carts, opts.BotID, senderID(c)), cartMarkup(t.carts, opts.BotID, senderID(c)))
}

// checkout moves the cart into a persisted order and clears the cart.
func (t *EcommerceShop) checkout(bot *tele.Bot, opts Options, sender *tele.User) (*models.Order, error) {
	if sender == nil {
		return nil, errEmptyCart
	}
	items, total := t.carts.snapshot(opts.BotID, sender.ID)
	if len(items) == 0 {
		return nil, errEmptyCart
	}
	order := &models.Order{
		BotID:      opts.BotID,
		TelegramID: sender.ID,
		Username:   sender.Username,
		Items:      items,
		Total:      total,
		Status:     models.OrderPending,
	}
	if err := opts.Store.CreateOrder(order); err != nil {
		return nil, err
	}
	t.carts.clear(opts.BotID, sender.ID)

	notifyAdmins(bot, opts, fmt.Sprintf("🛒 <b>Yangi buyurtma</b> #%d\n\n%sJami: <b>%s</b>\n👤 %s",
		order.ID, orderItemsText(order), money(order.Total)+" so'm", userLabel(sender)))
	return order, nil
}

// markPaid transitions an order owned by the user to paid.
func (t *EcommerceShop) markPaid(opts Options, userID, orderID int64) (*models.Order, error) {
	orders, err := opts.Store.ListOrdersByUser(opts.BotID, userID)
	if err != nil {
		return nil, err
	}
	for i := range orders {
		if orders[i].ID != orderID {
			continue
		}
		if orders[i].Status == models.OrderPaid {
			return &orders[i], nil
		}
		if err := opts.Store.UpdateOrderStatus(orderID, models.OrderPaid); err != nil {
			return nil, err
		}
		orders[i].Status = models.OrderPaid
		return &orders[i], nil
	}
	return nil, errOrderNotFound
}

func userLabel(u *tele.User) string {
	if u.Username != "" {
		return "@" + u.Username
	}
	name := strings.TrimSpace(u.FirstName + " " + u.LastName)
	if name != "" {
		return esc(name)
	}
	return strconv.FormatInt(u.ID, 10)
}

func catalogText() string {
	var b strings.Builder
	b.WriteString("🛍 <b>Do'kon katalogi</b>\n\n")
	for i, p := range ecommerceCatalog {
		fmt.Fprintf(&b, "<b>%d.</b> <b>%s</b> — %s\n", i+1, esc(p.Title), money(p.Price)+" so'm")
		if p.Description != "" {
			fmt.Fprintf(&b, "   %s\n", esc(p.Description))
		}
	}
	b.WriteString("\nMahsulot qo'shish uchun tugmani bosing. Savatcha: <code>/cart</code>")
	return b.String()
}

func catalogMarkup(cartCount int) *tele.ReplyMarkup {
	rows := make([][]tele.InlineButton, 0, len(ecommerceCatalog)+1)
	for _, p := range ecommerceCatalog {
		rows = append(rows, []tele.InlineButton{{
			Text: fmt.Sprintf("➕ %s — %s", p.Title, money(p.Price)+" so'm"),
			Data: "add_" + p.Code,
		}})
	}
	cartLabel := "🛒 Savatcha"
	if cartCount > 0 {
		cartLabel = fmt.Sprintf("🛒 Savatcha (%d)", cartCount)
	}
	rows = append(rows, []tele.InlineButton{{Text: cartLabel, Data: "cart"}})
	return &tele.ReplyMarkup{InlineKeyboard: rows}
}

func cartText(cs *cartState, botID, userID int64) string {
	items, total := cs.snapshot(botID, userID)
	if len(items) == 0 {
		return "🛒 Savatcha bo'sh."
	}
	var b strings.Builder
	b.WriteString("🛒 <b>Savatcha</b>\n\n")
	for i, it := range items {
		fmt.Fprintf(&b, "<b>%d.</b> <b>%s</b>\n", i+1, esc(it.Title))
		fmt.Fprintf(&b, "    %d × %s = <b>%s</b>\n", it.Qty, money(it.Price)+" so'm", money(it.Price*int64(it.Qty))+" so'm")
	}
	fmt.Fprintf(&b, "\nJami: <b>%s</b>", money(total)+" so'm")
	return b.String()
}

func cartMarkup(cs *cartState, botID, userID int64) *tele.ReplyMarkup {
	items, _ := cs.snapshot(botID, userID)
	rows := make([][]tele.InlineButton, 0, len(items)+2)
	for _, it := range items {
		rows = append(rows, []tele.InlineButton{
			{Text: "➖", Data: "dec_" + it.Code},
			{Text: fmt.Sprintf("%s × %d", esc(it.Title), it.Qty), Data: "noop"},
			{Text: "➕", Data: "inc_" + it.Code},
		})
	}
	rows = append(rows,
		[]tele.InlineButton{{Text: "✅ Rasmiylashtirish", Data: "checkout"}},
		[]tele.InlineButton{{Text: "🧹 Tozalash", Data: "clear"}, {Text: "🛍 Do'kon", Data: "menu"}},
	)
	return &tele.ReplyMarkup{InlineKeyboard: rows}
}

func orderItemsText(o *models.Order) string {
	var b strings.Builder
	for i, it := range o.Items {
		fmt.Fprintf(&b, "<b>%d.</b> <b>%s</b> — %d × %s\n",
			i+1, esc(it.Title), it.Qty, money(it.Price)+" so'm")
	}
	return b.String()
}

func orderText(o *models.Order) string {
	return fmt.Sprintf(`✅ <b>Buyurtma qabul qilindi!</b> #%d

%sJami: <b>%s</b>

💳 To'lovni ushbu kartaga o'tkazing:
<code>9860 1001 2345 6789</code> (Aliyev A.)

To'lovni tasdiqlash: <code>/pay %d</code> yoki tugmani bosing.`,
		o.ID, orderItemsText(o), money(o.Total)+" so'm", o.ID)
}

func orderPaidText(o *models.Order) string {
	return fmt.Sprintf(`✅ <b>To'lov tasdiqlandi!</b> #%d

%sJami: <b>%s</b>

📦 Operator tez orada siz bilan bog'lanadi. Rahmat!`,
		o.ID, orderItemsText(o), money(o.Total)+" so'm")
}

func payMarkup(o *models.Order) *tele.ReplyMarkup {
	return &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
		{{Text: "💳 To'landi", Data: fmt.Sprintf("pay_%d", o.ID)}},
		{{Text: "🛍 Do'kon", Data: "menu"}},
	}}
}

func menuOnlyMarkup() *tele.ReplyMarkup {
	return &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{
		{{Text: "🛍 Do'kon", Data: "menu"}},
	}}
}

func orderListText(orders []models.Order) string {
	var b strings.Builder
	b.WriteString("📦 <b>Buyurtmalarim</b>\n\n")
	for _, o := range orders {
		status := "⏳ Kutilmoqda"
		if o.Status == models.OrderPaid {
			status = "✅ To'langan"
		}
		fmt.Fprintf(&b, "#%d — <b>%s</b> — %s (%s)\n",
			o.ID, money(o.Total)+" so'm", status, o.CreatedAt.Format("02.01.2006 15:04"))
	}
	return b.String()
}

// money formats an integer amount in so'm with thousands separators.
func money(v int64) string {
	s := strconv.FormatInt(v, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + " " + s[i:]
	}
	return s
}

// ecommerceCatalog is the demo product dataset of the shop bot.
var ecommerceCatalog = []Product{
	{
		Code: "ECO-001", Title: "iPhone 15 Pro 256GB", Price: 12_900_000,
		Description: "Titanium korpus, A17 Pro chip, 48MP kamera", InStock: 10,
	},
	{
		Code: "ECO-002", Title: "Samsung Galaxy S24 Ultra", Price: 11_500_000,
		Description: "200MP kamera, S Pen, Snapdragon 8 Gen 3", InStock: 8,
	},
	{
		Code: "ECO-003", Title: "MacBook Air M3 13\"", Price: 14_200_000,
		Description: "8/256GB, 18 soatgacha batareya", InStock: 5,
	},
	{
		Code: "ECO-004", Title: "AirPods Pro 2", Price: 1_950_000,
		Description: "Faol shovqin o'chirish, USB-C", InStock: 25,
	},
	{
		Code: "ECO-005", Title: "Nike Air Force 1", Price: 1_100_000,
		Description: "Klassik krossovka, 41-45 o'lchamlar", InStock: 40,
	},
	{
		Code: "ECO-006", Title: "Adidas Tiro 24 (forma)", Price: 450_000,
		Description: "Futbol formasi, M-L-XL o'lchamlar", InStock: 60,
	},
	{
		Code: "ECO-007", Title: "Xiaomi Robot Vacuum S10", Price: 2_300_000,
		Description: "Aqlli changyutgich, uy xaritasi, 4000Pa", InStock: 12,
	},
	{
		Code: "ECO-008", Title: "Sony WH-1000XM5", Price: 3_100_000,
		Description: "Premium quloqchin, 30 soat batareya", InStock: 7,
	},
	{
		Code: "ECO-009", Title: "PlayStation 5 Slim", Price: 5_400_000,
		Description: "1TB SSD, DualSense gamepad", InStock: 6,
	},
	{
		Code: "ECO-010", Title: "LG 55\" OLED TV", Price: 7_800_000,
		Description: "4K OLED, 120Hz, Smart TV", InStock: 4,
	},
	{
		Code: "ECO-011", Title: "Samsung Galaxy Watch 6", Price: 2_150_000,
		Description: "Sog'liqni kuzatish, GPS, 44mm", InStock: 15,
	},
	{
		Code: "ECO-012", Title: "Kitob: Dasturlash asoslari", Price: 120_000,
		Description: "Python asoslari, yangi boshlanuvchilar uchun", InStock: 100,
	},
}
