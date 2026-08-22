package templates

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/internal/storage"
)

// ChannelManager is the channel management template: admins configure
// required channels, users must subscribe to them, and join requests to
// the connected channels are auto-approved.
type ChannelManager struct{}

// chatByUsername is a Recipient addressing a chat by its public @username.
type chatByUsername string

func (u chatByUsername) Recipient() string {
	return "@" + strings.TrimPrefix(string(u), "@")
}

// settingChannels is the bot_settings key holding the channel list.
const settingChannels = "channels"

// Name returns the stable template identifier.
func (t *ChannelManager) Name() string { return "channel_manager" }

// Title returns the display name of the template.
func (t *ChannelManager) Title() string { return "Channel Manager" }

// Description returns the human readable summary used by the API.
func (t *ChannelManager) Description() string {
	return "Kanal boshqaruvi — majburiy obuna tekshiruvi va join requestlarni avto-qabul qilish"
}

// Commands lists the commands exposed by this template.
func (t *ChannelManager) Commands() []string {
	return []string{
		"/start", "/help",
		"/channels",
		"/setchannels <@ch1 @ch2>",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the ChannelManager template on the bot.
func (t *ChannelManager) Apply(bot *tele.Bot, opts Options) error {
	applyCommon(bot, "Channel Manager", opts)

	bot.Handle("/start", func(c tele.Context) error {
		sender := c.Sender()
		if sender == nil {
			return nil
		}
		channels := loadChannels(opts.Store, opts.BotID)
		if len(channels) == 0 {
			return c.Send("⚠️ Majburiy kanallar hali sozlanmagan.\n\nAdmin: <code>/setchannels @kanal1 @kanal2</code>")
		}
		missing, ok := checkSubscription(bot, channels, sender.ID)
		if ok {
			return c.Send(welcomeText(sender))
		}
		return c.Send(subscriptionRequiredText(channels, missing), subscriptionMarkup(channels))
	})

	bot.Handle("/help", func(c tele.Context) error {
		return c.Send(`📢 <b>Channel Manager — yordam</b>

• <code>/start</code> — obunani tekshirish
• <code>/channels</code> — majburiy kanallar ro'yxati

🛡️ Adminlar uchun:
• <code>/setchannels @ch1 @ch2</code> — majburiy kanallarni o'rnatish
• Join requestlar avtomatik qabul qilinadi
• <code>/broadcast &lt;matn&gt;</code> — barchaga xabar`)
	})

	bot.Handle("/channels", func(c tele.Context) error {
		channels := loadChannels(opts.Store, opts.BotID)
		if len(channels) == 0 {
			return c.Send("📢 Majburiy kanallar sozlanmagan.")
		}
		var b strings.Builder
		b.WriteString("📢 <b>Majburiy kanallar:</b>\n\n")
		for i, ch := range channels {
			fmt.Fprintf(&b, "<b>%d.</b> %s\n", i+1, esc(ch))
		}
		return c.Send(b.String())
	})

	bot.Handle("/setchannels", func(c tele.Context) error {
		if !isAdmin(opts, c) {
			return c.Send("⛔ Bu buyruq faqat adminlar uchun.")
		}
		fields := strings.Fields(c.Data())
		if len(fields) == 0 {
			return c.Send("ℹ️ Foydalanish: <code>/setchannels @kanal1 @kanal2</code>")
		}
		seen := make(map[string]bool)
		var channels []string
		for _, f := range fields {
			f = strings.TrimSpace(f)
			if !strings.HasPrefix(f, "@") {
				f = "@" + f
			}
			if len(f) <= 1 || seen[f] {
				continue
			}
			seen[f] = true
			channels = append(channels, f)
		}
		if err := saveChannels(opts.Store, opts.BotID, channels); err != nil {
			return c.Send("❌ Kanallarni saqlab bo'lmadi.")
		}
		var b strings.Builder
		b.WriteString("✅ Majburiy kanallar saqlandi:\n\n")
		for i, ch := range channels {
			fmt.Fprintf(&b, "<b>%d.</b> %s\n", i+1, esc(ch))
		}
		return c.Send(b.String())
	})

	bot.Handle(tele.OnChatJoinRequest, func(c tele.Context) error {
		req := c.ChatJoinRequest()
		if req == nil || req.Chat == nil || req.Sender == nil {
			return nil
		}
		if err := bot.ApproveJoinRequest(req.Chat, req.Sender); err != nil {
			log.Printf("[bot %d] approve join request: %v", opts.BotID, err)
			return nil
		}
		if req.UserChatID != 0 {
			_, _ = bot.Send(tele.ChatID(req.UserChatID),
				"✅ Obuna so'rovingiz qabul qilindi. Xush kelibsiz!\n\nBotga qaytib <code>/start</code> ni bosing.")
		}
		return nil
	})

	bot.Handle(tele.OnCallback, func(c tele.Context) error {
		if c.Data() != "check_sub" {
			_ = c.Respond()
			return nil
		}
		sender := c.Sender()
		if sender == nil {
			return nil
		}
		channels := loadChannels(opts.Store, opts.BotID)
		if len(channels) == 0 {
			_ = c.Respond(&tele.CallbackResponse{Text: "Kanallar sozlanmagan", ShowAlert: true})
			return nil
		}
		missing, ok := checkSubscription(bot, channels, sender.ID)
		if ok {
			_ = c.Respond(&tele.CallbackResponse{Text: "✅ Obuna tasdiqlandi"})
			return c.Edit(welcomeText(sender))
		}
		_ = c.Respond(&tele.CallbackResponse{Text: "❌ Hali obuna bo'lmagansiz"})
		return c.Edit(subscriptionRequiredText(channels, missing), subscriptionMarkup(channels))
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("ℹ️ Obunani tekshirish uchun <code>/start</code> ni bosing.")
	})
	return nil
}

// loadChannels reads the configured channel list of a bot.
func loadChannels(store *storage.Store, botID int64) []string {
	raw, err := store.GetSetting(botID, settingChannels)
	if err != nil {
		return nil
	}
	var channels []string
	if json.Unmarshal([]byte(raw), &channels) != nil {
		return nil
	}
	return channels
}

// saveChannels persists the channel list of a bot.
func saveChannels(store *storage.Store, botID int64, channels []string) error {
	data, err := json.Marshal(channels)
	if err != nil {
		return err
	}
	return store.SetSetting(botID, settingChannels, string(data))
}

// checkSubscription verifies the user is a member of every channel and
// returns the list of channels the user still misses.
func checkSubscription(bot *tele.Bot, channels []string, userID int64) (missing []string, ok bool) {
	for _, ch := range channels {
		if !isSubscribed(bot, ch, userID) {
			missing = append(missing, ch)
		}
	}
	return missing, len(missing) == 0
}

func isSubscribed(bot *tele.Bot, username string, userID int64) bool {
	member, err := bot.ChatMemberOf(chatByUsername(username), tele.ChatID(userID))
	if err != nil {
		return false
	}
	switch member.Role {
	case tele.Creator, tele.Administrator, tele.Member:
		return true
	case tele.Restricted:
		return member.Member
	}
	return false
}

func subscriptionMarkup(channels []string) *tele.ReplyMarkup {
	rows := make([][]tele.InlineButton, 0, len(channels)+1)
	for _, ch := range channels {
		name := strings.TrimPrefix(ch, "@")
		rows = append(rows, []tele.InlineButton{{
			Text: "📢 " + ch,
			URL:  "https://t.me/" + name,
		}})
	}
	rows = append(rows, []tele.InlineButton{{Text: "✅ Tekshirish", Data: "check_sub"}})
	return &tele.ReplyMarkup{InlineKeyboard: rows}
}

func subscriptionRequiredText(channels, missing []string) string {
	var b strings.Builder
	b.WriteString("🔒 <b>Majburiy obuna</b>\n\n")
	b.WriteString("Botdan foydalanish uchun quyidagi kanallarga obuna bo'ling:\n\n")
	for i, ch := range channels {
		state := "❌"
		if !containsString(missing, ch) {
			state = "✅"
		}
		fmt.Fprintf(&b, "%s <b>%d.</b> %s\n", state, i+1, esc(ch))
	}
	b.WriteString("\nObuna bo'lgach <b>\"✅ Tekshirish\"</b> tugmasini bosing.")
	return b.String()
}

func welcomeText(sender *tele.User) string {
	name := sender.FirstName
	if name == "" {
		name = sender.Username
	}
	return fmt.Sprintf("🎉 Xush kelibsiz, <b>%s</b>!\n\n✅ Barcha kanallarga obuna bo'lgansiz. Botdan bemalol foydalanishingiz mumkin.", esc(name))
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
