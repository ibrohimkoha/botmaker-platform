package templates

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/internal/storage"
)

// CustomBuilder is the visual menu template: admins define buttons and
// their replies, users tap a button and get the configured response.
type CustomBuilder struct{}

// menuItem is one button of the custom menu.
type menuItem struct {
	Label string `json:"label"`
	Reply string `json:"reply"`
}

// settingMenu is the bot_settings key holding the menu definition.
const settingMenu = "menu"

// maxMenuItems bounds the size of the menu (Telegram keyboard limits).
const maxMenuItems = 50

// Name returns the stable template identifier.
func (t *CustomBuilder) Name() string { return "custom_builder" }

// Title returns the display name of the template.
func (t *CustomBuilder) Title() string { return "Custom Builder" }

// Description returns the human readable summary used by the API.
func (t *CustomBuilder) Description() string {
	return "Vizual menyu va tugmalar boti — sozlanuvchi tugmalar va javoblar"
}

// Commands lists the commands exposed by this template.
func (t *CustomBuilder) Commands() []string {
	return []string{
		"/start", "/help",
		"/menu",
		"/addmenu <label> | <javob>",
		"/delmenu <label>",
		"/clearmenu",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the CustomBuilder template on the bot.
func (t *CustomBuilder) Apply(bot *tele.Bot, opts Options) error {
	applyCommon(bot, "Custom Builder", opts)

	bot.Handle("/start", func(c tele.Context) error {
		name := ""
		if sender := c.Sender(); sender != nil {
			name = sender.FirstName
			if name == "" {
				name = sender.Username
			}
		}
		return c.Send(fmt.Sprintf("👋 Salom, <b>%s</b>!\n\n🎛 <b>Custom Builder</b> — vizual menyu boti.\n\nQuyidagi tugmalardan birini tanlang yoki <code>/menu</code> ni bosing.", esc(name)),
			menuMarkup(loadMenu(opts.Store, opts.BotID)))
	})

	bot.Handle("/help", func(c tele.Context) error {
		return c.Send(`🎛 <b>Custom Builder — yordam</b>

• <code>/menu</code> — menyuni ko'rish
• Tugmani bosib javob olish

🛡️ Adminlar uchun:
• <code>/addmenu &lt;label&gt; | &lt;javob&gt;</code> — tugma qo'shish
• <code>/delmenu &lt;label&gt;</code> — tugmani o'chirish
• <code>/clearmenu</code> — barcha tugmalarni tozalash
• <code>/broadcast &lt;matn&gt;</code> — barchaga xabar`)
	})

	bot.Handle("/menu", func(c tele.Context) error {
		items := loadMenu(opts.Store, opts.BotID)
		if len(items) == 0 {
			return c.Send("🎛 Menyu hali bo'sh. Admin: <code>/addmenu Tugma | Javob</code>")
		}
		return c.Send("🎛 <b>Menyu:</b>\n\nTugmani bosing:", menuMarkup(items))
	})

	bot.Handle("/addmenu", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		raw := strings.TrimSpace(c.Data())
		if raw == "" {
			return c.Send("ℹ️ Foydalanish: <code>/addmenu Tugma nomi | Javob matni</code>")
		}
		parts := strings.SplitN(raw, "|", 2)
		label := strings.TrimSpace(parts[0])
		reply := strings.TrimSpace(parts[1])
		if label == "" {
			return c.Send("ℹ️ Tugma nomini kiriting: <code>/addmenu Tugma | Javob</code>")
		}

		items := loadMenu(opts.Store, opts.BotID)
		if len(items) >= maxMenuItems {
			return c.Send(fmt.Sprintf("⚠️ Menyu sig'imi to'ldi (maksimum %d ta tugma).", maxMenuItems))
		}
		for i := range items {
			if items[i].Label == label {
				items[i].Reply = reply
				if err := saveMenu(opts.Store, opts.BotID, items); err != nil {
					return c.Send("❌ Saqlab bo'lmadi.")
				}
				return c.Send(fmt.Sprintf("✅ Tugma yangilandi: <b>%s</b>", esc(label)))
			}
		}
		items = append(items, menuItem{Label: label, Reply: reply})
		if err := saveMenu(opts.Store, opts.BotID, items); err != nil {
			return c.Send("❌ Saqlab bo'lmadi.")
		}
		return c.Send(fmt.Sprintf("✅ Tugma qo'shildi: <b>%s</b>\n\nMenyuni ko'rish: /menu", esc(label)))
	})

	bot.Handle("/delmenu", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		label := strings.TrimSpace(c.Data())
		if label == "" {
			return c.Send("ℹ️ Foydalanish: <code>/delmenu Tugma nomi</code>")
		}
		items := loadMenu(opts.Store, opts.BotID)
		filtered := items[:0]
		for _, it := range items {
			if it.Label != label {
				filtered = append(filtered, it)
			}
		}
		if len(filtered) == len(items) {
			return c.Send("😕 Bunday tugma topilmadi.")
		}
		if err := saveMenu(opts.Store, opts.BotID, filtered); err != nil {
			return c.Send("❌ Saqlab bo'lmadi.")
		}
		return c.Send(fmt.Sprintf("🗑 Tugma o'chirildi: <b>%s</b>", esc(label)))
	})

	bot.Handle("/clearmenu", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		if err := opts.Store.DeleteSetting(opts.BotID, settingMenu); err != nil {
			return c.Send("❌ Menyu tozalanmadi.")
		}
		return c.Send("🧹 Menyu tozalandi.")
	})

	bot.Handle(tele.OnCallback, func(c tele.Context) error {
		data := c.Data()
		if !strings.HasPrefix(data, "menu_") {
			_ = c.Respond()
			return nil
		}
		idx, err := strconv.Atoi(strings.TrimPrefix(data, "menu_"))
		items := loadMenu(opts.Store, opts.BotID)
		if err != nil || idx < 0 || idx >= len(items) {
			_ = c.Respond(&tele.CallbackResponse{Text: "Menyu yangilangan", ShowAlert: true})
			return nil
		}
		it := items[idx]
		if it.Reply != "" {
			_ = c.Respond()
			return c.Send(esc(it.Reply))
		}
		_ = c.Respond(&tele.CallbackResponse{Text: it.Label})
		return nil
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		if isAdmin(opts, c) {
			return c.Send("ℹ️ Tugma qo'shish: <code>/addmenu Tugma | Javob</code>. Menyu: /menu")
		}
		return c.Send("🎛 Menyudan tugma tanlang: /menu")
	})
	return nil
}

// loadMenu reads the custom menu definition of a bot.
func loadMenu(store *storage.Store, botID int64) []menuItem {
	raw, err := store.GetSetting(botID, settingMenu)
	if err != nil {
		return nil
	}
	var items []menuItem
	if json.Unmarshal([]byte(raw), &items) != nil {
		return nil
	}
	return items
}

// saveMenu persists the custom menu definition of a bot.
func saveMenu(store *storage.Store, botID int64, items []menuItem) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return store.SetSetting(botID, settingMenu, string(data))
}

// menuMarkup renders the menu as an inline keyboard, two buttons per row.
func menuMarkup(items []menuItem) *tele.ReplyMarkup {
	if len(items) == 0 {
		return nil
	}
	rows := make([][]tele.InlineButton, 0, (len(items)+1)/2)
	for i := 0; i < len(items); i += 2 {
		row := []tele.InlineButton{{
			Text: items[i].Label,
			Data: "menu_" + strconv.Itoa(i),
		}}
		if i+1 < len(items) {
			row = append(row, tele.InlineButton{
				Text: items[i+1].Label,
				Data: "menu_" + strconv.Itoa(i+1),
			})
		}
		rows = append(rows, row)
	}
	return &tele.ReplyMarkup{InlineKeyboard: rows}
}
