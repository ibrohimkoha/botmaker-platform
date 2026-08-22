package templates

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/internal/models"
)

// FeedbackSupport is the contact & support template: user messages are
// forwarded to every admin, who can reply inline or via /reply.
type FeedbackSupport struct {
	state *supportState
}

// supportState maps forwarded messages in admin chats back to the user
// that sent them, so a reply-to hits the right person.
type supportState struct {
	mu        sync.Mutex
	forwarded map[int64]map[int64]map[int]int64 // botID -> adminID -> msgID -> userID
}

func (s *supportState) remember(botID, adminID int64, msgID int, userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.forwarded[botID] == nil {
		s.forwarded[botID] = make(map[int64]map[int]int64)
	}
	if s.forwarded[botID][adminID] == nil {
		s.forwarded[botID][adminID] = make(map[int]int64)
	}
	s.forwarded[botID][adminID][msgID] = userID
}

func (s *supportState) lookup(botID, adminID int64, msgID int) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	uid, ok := s.forwarded[botID][adminID][msgID]
	return uid, ok
}

// Name returns the stable template identifier.
func (t *FeedbackSupport) Name() string { return "feedback_support" }

// Title returns the display name of the template.
func (t *FeedbackSupport) Title() string { return "Feedback & Support" }

// Description returns the human readable summary used by the API.
func (t *FeedbackSupport) Description() string {
	return "Bog'lanish boti — foydalanuvchi xabarlari adminga yetkaziladi, admin javob qaytaradi"
}

// Commands lists the commands exposed by this template.
func (t *FeedbackSupport) Commands() []string {
	return []string{
		"/start", "/help",
		"/reply <id> <matn>",
		"/tickets",
		"/close <id>",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the FeedbackSupport template on the bot.
func (t *FeedbackSupport) Apply(bot *tele.Bot, opts Options) error {
	if t.state == nil {
		t.state = &supportState{forwarded: make(map[int64]map[int64]map[int]int64)}
	}

	applyCommon(bot, "Feedback & Support", opts)

	bot.Handle("/start", func(c tele.Context) error {
		name := ""
		if sender := c.Sender(); sender != nil {
			name = sender.FirstName
			if name == "" {
				name = sender.Username
			}
		}
		return c.Send(fmt.Sprintf("👋 Salom, <b>%s</b>!\n\n💬 Savol, taklif yoki muammoingizni yozing — xabaringiz adminga yetkaziladi va siz javob olasiz.", esc(name)))
	})

	bot.Handle("/help", func(c tele.Context) error {
		return c.Send(`💬 <b>Feedback & Support — yordam</b>

• Savol yoki muammoingizni ochiq matnda yozing
• Admin javobini shu chatda olasiz

🛡️ Adminlar uchun:
• Xabarga reply qilib javob berish
• <code>/reply &lt;id&gt; &lt;matn&gt;</code> — foydalanuvchiga javob
• <code>/tickets</code> — ochiq murojaatlar
• <code>/close &lt;id&gt;</code> — murojaatni yopish
• <code>/broadcast &lt;matn&gt;</code> — barchaga xabar`)
	})

	bot.Handle("/reply", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		data := strings.TrimSpace(c.Data())
		fields := strings.Fields(data)
		if len(fields) < 2 {
			return c.Send("ℹ️ Foydalanish: <code>/reply &lt;user_id&gt; &lt;matn&gt;</code>")
		}
		uid, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return c.Send("ℹ️ Foydalanish: <code>/reply &lt;user_id&gt; &lt;matn&gt;</code>")
		}
		text := strings.TrimSpace(data[len(fields[0]):])
		if text == "" {
			return c.Send("ℹ️ Javob matnini kiriting: <code>/reply &lt;user_id&gt; &lt;matn&gt;</code>")
		}
		if _, err := bot.Send(tele.ChatID(uid), "📩 <b>Admin javobi:</b>\n"+esc(text)); err != nil {
			return c.Send("❌ Foydalanuvchiga yuborib bo'lmadi.")
		}
		return c.Send("✅ Javob foydalanuvchiga yuborildi.")
	})

	bot.Handle("/tickets", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		tickets, err := opts.Store.ListTickets(opts.BotID)
		if err != nil {
			return c.Send("❌ Murojaatlarni o'qib bo'lmadi.")
		}
		var open []models.Ticket
		for _, tk := range tickets {
			if tk.Status == models.TicketOpen {
				open = append(open, tk)
			}
		}
		if len(open) == 0 {
			return c.Send("🎫 Ochiq murojaatlar yo'q.")
		}
		var b strings.Builder
		b.WriteString("🎫 <b>Ochiq murojaatlar</b>\n\n")
		for _, tk := range open {
			fmt.Fprintf(&b, "<b>%d.</b> %s (id: <code>%d</code>) — %s\n",
				tk.ID, userLabel(&tele.User{ID: tk.TelegramID, Username: tk.Username}),
				tk.TelegramID, tk.UpdatedAt.Format("02.01.2006 15:04"))
		}
		b.WriteString("\nJavob: <code>/reply &lt;id&gt; &lt;matn&gt;</code>")
		return c.Send(b.String())
	})

	bot.Handle("/close", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		uid, err := strconv.ParseInt(strings.TrimSpace(c.Data()), 10, 64)
		if err != nil {
			return c.Send("ℹ️ Foydalanish: <code>/close &lt;user_id&gt;</code>")
		}
		if err := opts.Store.SetTicketStatus(opts.BotID, uid, models.TicketClosed); err != nil {
			return c.Send("❌ Murojaatni yopib bo'lmadi.")
		}
		_, _ = bot.Send(tele.ChatID(uid), "✅ Murojaatingiz yopildi. Yana savol bo'lsa, yozishingiz mumkin.")
		return c.Send("✅ Murojaat yopildi.")
	})

	incoming := func(c tele.Context) error { return t.handleIncoming(bot, opts, c) }
	bot.Handle(tele.OnText, incoming)
	bot.Handle(tele.OnPhoto, incoming)
	bot.Handle(tele.OnVideo, incoming)
	bot.Handle(tele.OnDocument, incoming)
	bot.Handle(tele.OnVoice, incoming)
	bot.Handle(tele.OnAnimation, incoming)
	return nil
}

// handleIncoming routes a message: admin replies are sent back to the
// user, everything else is forwarded to every admin.
func (t *FeedbackSupport) handleIncoming(bot *tele.Bot, opts Options, c tele.Context) error {
	sender := c.Sender()
	m := c.Message()
	if sender == nil || m == nil {
		return nil
	}

	if isAdmin(opts, c) {
		if m.ReplyTo != nil {
			uid, ok := t.state.lookup(opts.BotID, sender.ID, m.ReplyTo.ID)
			if !ok {
				return c.Send("😕 Bu xabar bo'yicha foydalanuvchi topilmadi.")
			}
			text := strings.TrimSpace(c.Text())
			if text == "" {
				return c.Send("ℹ️ Javob matnini kiriting.")
			}
			if _, err := bot.Send(tele.ChatID(uid), "📩 <b>Admin javobi:</b>\n"+esc(text)); err != nil {
				return c.Send("❌ Foydalanuvchiga yuborib bo'lmadi.")
			}
			return c.Send("✅ Javob foydalanuvchiga yuborildi.")
		}
		return c.Send("ℹ️ Javob berish uchun foydalanuvchi xabariga <b>reply</b> qiling yoki <code>/reply &lt;id&gt; &lt;matn&gt;</code> dan foydalaning.")
	}

	// Regular user: forward to every admin and open a ticket.
	for _, adminID := range opts.AdminIDs {
		fwd, err := bot.Forward(tele.ChatID(adminID), m)
		if err != nil {
			log.Printf("[bot %d] forward to admin %d: %v", opts.BotID, adminID, err)
			continue
		}
		t.state.remember(opts.BotID, adminID, fwd.ID, sender.ID)
		_, _ = bot.Send(tele.ChatID(adminID), fmt.Sprintf(
			"👤 <b>%s</b> (id: <code>%d</code>)\nJavob: <code>/reply %d &lt;matn&gt;</code> yoki xabarga reply qiling",
			userLabel(sender), sender.ID, sender.ID))
	}

	_ = opts.Store.UpsertTicket(&models.Ticket{
		BotID:      opts.BotID,
		TelegramID: sender.ID,
		Username:   sender.Username,
		Status:     models.TicketOpen,
	})
	return c.Send("✅ Xabaringiz adminga yetkazildi. Javob kelishini kuting.")
}
