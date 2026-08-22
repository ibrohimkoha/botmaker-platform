package templates

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/internal/models"
	"botmaker-backend/internal/storage"
)

// Template describes a bot behaviour template that can be applied to a
// Telegram bot instance.
type Template interface {
	Name() string
	Title() string
	Description() string
	Commands() []string
	Apply(bot *tele.Bot, opts Options) error
}

// Options carries per-bot dependencies into a template.
type Options struct {
	BotID    int64
	AdminIDs []int64
	Store    *storage.Store
}

// Spec describes the concrete behaviour of a single template.
type Spec struct {
	// ID is the stable template identifier (e.g. "anitez").
	ID string
	// Name is the display name of the bot (e.g. "AniTez").
	Name string
	// Tagline is shown on /start below the greeting.
	Tagline string
	// SearchNoun describes what can be searched (e.g. "anime va kino").
	SearchNoun string
	// CodePrefix is the prefix of every catalog code (e.g. "ANZ").
	CodePrefix string
	// Catalog is the searchable dataset of the bot.
	Catalog []Title
	// ExtraCmds are additional command lines appended to /help.
	ExtraCmds []string
}

// codeRx matches catalog codes like "ANZ-001".
var codeRx = regexp.MustCompile(`^[A-Za-z]{2,4}-\d{1,5}$`)

func esc(s string) string {
	return html.EscapeString(s)
}

func isAdmin(opts Options, c tele.Context) bool {
	sender := c.Sender()
	if sender == nil {
		return false
	}
	for _, id := range opts.AdminIDs {
		if id == sender.ID {
			return true
		}
	}
	return false
}

// applyCore registers the handlers shared by every template: start,
// help, search, code lookup, stats, admin panel, broadcast, inline
// search and the plain-text fallback.
func applyCore(bot *tele.Bot, spec Spec, opts Options) {
	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			trackUser(opts, c)
			return next(c)
		}
	})

	bot.Handle("/start", func(c tele.Context) error {
		name := ""
		if sender := c.Sender(); sender != nil {
			name = sender.FirstName
			if name == "" {
				name = sender.Username
			}
		}
		msg := fmt.Sprintf("👋 Salom, <b>%s</b>!\n\n%s\n\n%s",
			esc(name), spec.Tagline, usageText(spec))
		if code := normalizeCode(c.Data()); code != "" {
			if t, ok := FindByCode(spec.Catalog, code); ok {
				msg += "\n\n🔎 So'rovingiz bo'yicha:\n" + formatCard(spec, t)
			}
		}
		return c.Send(msg)
	})

	bot.Handle("/help", func(c tele.Context) error {
		return c.Send(helpText(spec))
	})

	bot.Handle("/search", func(c tele.Context) error {
		q := strings.TrimSpace(c.Data())
		if q == "" {
			return c.Send(fmt.Sprintf("ℹ️ Qidiruv so'rovini kiriting, masalan: <code>/search %s</code>", esc(spec.ExampleQuery())))
		}
		_ = opts.Store.RecordSearch(opts.BotID)
		hits := SearchCatalog(spec.Catalog, q)
		if len(hits) == 0 {
			return c.Send(fmt.Sprintf("😕 <b>%s</b> bo'yicha hech narsa topilmadi.\n\nBoshqa so'rov bilan urinib ko'ring yoki /help ni bosing.", esc(q)))
		}
		var b strings.Builder
		fmt.Fprintf(&b, "🔎 <b>%s</b> — %d ta natija:\n\n", esc(q), len(hits))
		for i, t := range hits {
			if i >= 5 {
				fmt.Fprintf(&b, "\n… va yana %d ta natija. Batafsil: /code <KOD>", len(hits)-5)
				break
			}
			fmt.Fprintf(&b, "<b>%d.</b> <b>%s</b> [<code>%s</code>]\n", i+1, esc(t.Title), t.Code)
			fmt.Fprintf(&b, "    %s • %d • ⭐ %.1f\n", esc(t.TypeLabel()), t.Year, t.Rating)
		}
		return c.Send(b.String())
	})

	bot.Handle("/code", func(c tele.Context) error {
		code := normalizeCode(c.Data())
		if code == "" {
			return c.Send(fmt.Sprintf("ℹ️ Kod kiriting, masalan: <code>/code %s-001</code>", spec.CodePrefix))
		}
		_ = opts.Store.RecordCodeLookup(opts.BotID)
		t, ok := FindByCode(spec.Catalog, code)
		if !ok {
			return c.Send(fmt.Sprintf("😕 <code>%s</code> kodi topilmadi.\n\nTo'g'ri format: <code>/code %s-001</code>", esc(code), spec.CodePrefix))
		}
		return c.Send(formatCard(spec, t))
	})

	bot.Handle("/stats", func(c tele.Context) error {
		st, err := opts.Store.GetStats(opts.BotID)
		if err != nil {
			return c.Send("❌ Statistikani o'qib bo'lmadi.")
		}
		return c.Send(formatStats(spec, st))
	})

	bot.Handle("/admin", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		st, err := opts.Store.GetStats(opts.BotID)
		if err != nil {
			return c.Send("❌ Statistikani o'qib bo'lmadi.")
		}
		return c.Send(fmt.Sprintf("🛡️ <b>Admin panel</b>\n\n%s\n\n📣 Barcha foydalanuvchilarga xabar yuborish:\n<code>/broadcast &lt;matn&gt;</code>",
			formatStats(spec, st)))
	})

	bot.Handle("/broadcast", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		msg := strings.TrimSpace(c.Data())
		if msg == "" {
			return c.Send("ℹ️ Xabar matnini kiriting: <code>/broadcast &lt;matn&gt;</code>")
		}
		ids, err := opts.Store.ListUserTelegramIDs(opts.BotID)
		if err != nil {
			return c.Send("❌ Foydalanuvchilar ro'yxatini o'qib bo'lmadi.")
		}
		rec := &models.Broadcast{
			BotID:   opts.BotID,
			Message: msg,
			Status:  models.BroadcastRunning,
			Total:   len(ids),
		}
		if err := opts.Store.CreateBroadcast(rec); err != nil {
			return c.Send("❌ Broadcast yaratilmadi.")
		}
		go runBroadcast(bot, opts.Store, rec, ids)
		return c.Send(fmt.Sprintf("📣 Broadcast boshlandi: <b>%d</b> foydalanuvchiga yuborilmoqda…", len(ids)))
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		text := strings.TrimSpace(c.Text())
		if code := normalizeCode(text); codeRx.MatchString(text) {
			_ = opts.Store.RecordCodeLookup(opts.BotID)
			if t, ok := FindByCode(spec.Catalog, code); ok {
				return c.Send(formatCard(spec, t))
			}
		}
		return c.Send("ℹ️ Buyruq tan olinmadi.\n\n" + usageText(spec))
	})

	bot.Handle(tele.OnQuery, func(c tele.Context) error {
		_ = opts.Store.RecordSearch(opts.BotID)
		q := strings.TrimSpace(c.Query().Text)
		var out tele.Results
		if q == "" {
			for _, t := range TopRated(spec.Catalog, 10) {
				out = append(out, articleResult(spec, t))
			}
		} else {
			for _, t := range SearchCatalog(spec.Catalog, q) {
				if len(out) >= 10 {
					break
				}
				out = append(out, articleResult(spec, t))
			}
		}
		return c.Answer(&tele.QueryResponse{
			Results:    out,
			CacheTime:  60,
			IsPersonal: true,
		})
	})
}

// ExampleQuery returns a sample search term taken from the catalog.
func (s Spec) ExampleQuery() string {
	if len(s.Catalog) > 0 {
		return s.Catalog[0].Title
	}
	return "so'rov"
}

func trackUser(opts Options, c tele.Context) {
	if c.Message() != nil {
		_ = opts.Store.RecordMessage(opts.BotID)
	}
	_ = opts.Store.TouchActivity(opts.BotID)

	sender := c.Sender()
	if sender == nil {
		return
	}
	u := &models.User{
		BotID:      opts.BotID,
		TelegramID: sender.ID,
		Username:   sender.Username,
		FirstName:  sender.FirstName,
		LastName:   sender.LastName,
		IsAdmin:    isAdmin(opts, c),
	}
	_, _ = opts.Store.UpsertUser(u)
}

func runBroadcast(bot *tele.Bot, store *storage.Store, rec *models.Broadcast, ids []int64) {
	if len(ids) == 0 {
		rec.Status = models.BroadcastDone
		_ = store.UpdateBroadcast(rec)
		return
	}
	for _, id := range ids {
		if _, err := bot.Send(tele.ChatID(id), rec.Message); err != nil {
			rec.Failed++
		} else {
			rec.Sent++
		}
		if (rec.Sent+rec.Failed)%10 == 0 || rec.Sent+rec.Failed == len(ids) {
			_ = store.UpdateBroadcast(rec)
		}
		time.Sleep(50 * time.Millisecond)
	}
	rec.Status = models.BroadcastDone
	_ = store.UpdateBroadcast(rec)
}

func formatCard(spec Spec, t Title) string {
	return fmt.Sprintf(`🎬 <b>%s</b> [<code>%s</code>]
━━━━━━━━━━━━━━━━━━
📺 Turi: %s
📅 Yil: %d
⭐ Reyting: %.1f
%s🏷️ Janrlar: %s
📝 Izoh: %s

🔗 Qayta ochish: <code>/code %s</code>`,
		esc(t.Title), t.Code, esc(t.TypeLabel()), t.Year, t.Rating,
		t.EpisodeLine(), esc(strings.Join(t.Genres, ", ")), esc(t.Description), t.Code)
}

func formatStats(spec Spec, st *models.Stats) string {
	return fmt.Sprintf(`📊 <b>%s — statistika</b>

👥 Foydalanuvchilar: %d
🆕 Bugun kelganlar: %d
💬 Xabarlar: %d
🔎 Qidiruvlar: %d
🔢 Kod bo'yicha: %d
📢 Broadcastlar: %d`,
		esc(spec.Name), st.TotalUsers, st.NewUsersToday, st.TotalMessages,
		st.TotalSearches, st.TotalCodeLookups, st.BroadcastsSent)
}

func usageText(spec Spec) string {
	return fmt.Sprintf(`🔍 <b>Qanday qidirish mumkin?</b>
• /search <i>so'rov</i> — %s qidirish
• /code <i>KOD</i> — kod orqali topish (masalan: <code>/code %s-001</code>)
• /stats — statistika
• /help — to'liq yordam`, esc(spec.SearchNoun), spec.CodePrefix)
}

func helpText(spec Spec) string {
	lines := []string{
		fmt.Sprintf("🤖 <b>%s</b> — yordam", esc(spec.Name)),
		"",
		"🔍 <b>Qidiruv:</b>",
		fmt.Sprintf("• /search <i>so'rov</i> — %s qidirish", esc(spec.SearchNoun)),
		fmt.Sprintf("• /code <i>KOD</i> — kod orqali topish (masalan: <code>/code %s-001</code>)", spec.CodePrefix),
		"• Inline rejim: istalgan chatda @username yozib qidirish",
		"",
		"📊 <b>Boshqa:</b>",
		"• /stats — statistika",
		"• /help — yordam",
	}
	lines = append(lines, spec.ExtraCmds...)
	lines = append(lines, "", "🛡️ Adminlar uchun: /admin, /broadcast <matn>")
	return strings.Join(lines, "\n")
}

func articleResult(spec Spec, t Title) tele.Result {
	return &tele.ArticleResult{
		ResultBase: tele.ResultBase{
			ParseMode: tele.ModeHTML,
		},
		Title:       t.Title,
		Description: fmt.Sprintf("%s • %d • ⭐ %.1f", t.TypeLabel(), t.Year, t.Rating),
		Text:        formatCard(spec, t),
	}
}
