package templates

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/config"
)

// AIAssistant is the LLM chatbot template backed by any OpenAI-compatible
// chat API (DeepSeek by default). Conversations keep a short rolling
// history per (bot, user) so follow-up questions have context.
type AIAssistant struct {
	memory *chatMemory
}

// chatMsg is one exchange of the assistant conversation.
type chatMsg struct {
	Role    string `json:"role"` // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// chatMemory keeps a bounded conversation history per (bot, user).
type chatMemory struct {
	mu      sync.Mutex
	history map[int64]map[int64][]chatMsg // botID -> userID -> messages
	maxLen  int
}

func (m *chatMemory) get(botID, userID int64) []chatMsg {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]chatMsg(nil), m.history[botID][userID]...)
}

func (m *chatMemory) append(botID, userID int64, msgs ...chatMsg) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.history[botID] == nil {
		m.history[botID] = make(map[int64][]chatMsg)
	}
	hist := m.history[botID][userID]
	hist = append(hist, msgs...)
	if len(hist) > m.maxLen {
		hist = hist[len(hist)-m.maxLen:]
	}
	m.history[botID][userID] = hist
}

func (m *chatMemory) clear(botID, userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.history[botID], userID)
}

// Name returns the stable template identifier.
func (t *AIAssistant) Name() string { return "ai_assistant" }

// Title returns the display name of the template.
func (t *AIAssistant) Title() string { return "AI Assistant" }

// Description returns the human readable summary used by the API.
func (t *AIAssistant) Description() string {
	return "DeepSeek/OpenAI sun'iy intellekt boti — savollarga aqlli javob beradi"
}

// Commands lists the commands exposed by this template.
func (t *AIAssistant) Commands() []string {
	return []string{
		"/start", "/help",
		"/reset",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the AIAssistant template on the bot.
func (t *AIAssistant) Apply(bot *tele.Bot, opts Options) error {
	if t.memory == nil {
		t.memory = &chatMemory{history: make(map[int64]map[int64][]chatMsg), maxLen: 20}
	}

	applyCommon(bot, "AI Assistant", opts)

	bot.Handle("/start", func(c tele.Context) error {
		name := ""
		if sender := c.Sender(); sender != nil {
			name = sender.FirstName
			if name == "" {
				name = sender.Username
			}
		}
		return c.Send(fmt.Sprintf("👋 Salom, <b>%s</b>!\n\n🤖 Men sun'iy intellekt yordamchisiman. Menga istalgan savolingizni yozing — aqlli javob beraman.\n\n💡 Suhbatni qayta boshlash uchun: <code>/reset</code>", esc(name)))
	})

	bot.Handle("/help", func(c tele.Context) error {
		return c.Send(`🤖 <b>AI Assistant — yordam</b>

💬 Menga ochiq matnda savol yozing, men javob beraman.
• <code>/reset</code> — suhbat tarixini tozalash
• <code>/stats</code> — statistika
• <code>/help</code> — yordam

🛡️ Adminlar uchun: /admin, /broadcast <matn>`)
	})

	bot.Handle("/reset", func(c tele.Context) error {
		if sender := c.Sender(); sender != nil {
			t.memory.clear(opts.BotID, sender.ID)
		}
		return c.Send("🧹 Suhbat tarixi tozalandi. Yangi suhbatni boshlashimiz mumkin!")
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		sender := c.Sender()
		if sender == nil {
			return nil
		}
		if opts.AI.APIKey == "" {
			return c.Send("⚠️ AI hali sozlanmagan. Operator bilan bog'laning.")
		}
		text := strings.TrimSpace(c.Text())
		if text == "" {
			return nil
		}

		_ = bot.Notify(tele.ChatID(sender.ID), tele.Typing)

		hist := t.memory.get(opts.BotID, sender.ID)
		hist = append(hist, chatMsg{Role: "user", Content: text})
		answer, err := askAI(opts.AI, hist)
		if err != nil {
			return c.Send("😕 Javob olishda xatolik yuz berdi. Iltimos, birozdan so'ng qayta urinib ko'ring.")
		}

		t.memory.append(opts.BotID, sender.ID,
			chatMsg{Role: "user", Content: text},
			chatMsg{Role: "assistant", Content: answer})
		return c.Send(esc(truncateRunes(answer, 4000)))
	})
	return nil
}

// truncateRunes limits s to at most n runes (Telegram's 4096 char limit).
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// chatCompletionRequest mirrors the OpenAI chat completions schema.
type chatCompletionRequest struct {
	Model     string    `json:"model"`
	Messages  []chatMsg `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

// chatCompletionResponse is the subset of the OpenAI response we need.
type chatCompletionResponse struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
}

const aiSystemPrompt = "Siz BotMaker platformasidagi aqlli yordamchi botsiz. " +
	"Savollarga aniq, qisqa va foydali javob bering. O'zbek tilida javob bering " +
	"(agar foydalanuvchi boshqa til so'ramasa)."

// askAI sends the conversation to the OpenAI-compatible endpoint and
// returns the assistant's reply.
func askAI(ai config.AIConfig, history []chatMsg) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	msgs := make([]chatMsg, 0, len(history)+1)
	msgs = append(msgs, chatMsg{Role: "system", Content: aiSystemPrompt})
	msgs = append(msgs, history...)

	body, err := json.Marshal(chatCompletionRequest{
		Model:     ai.Model,
		Messages:  msgs,
		MaxTokens: 1024,
	})
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimSuffix(ai.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call AI: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI status %d: %s", resp.StatusCode, truncate(string(data), 300))
	}

	var out chatCompletionResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("AI returned no choices")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
