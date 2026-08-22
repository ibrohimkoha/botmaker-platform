package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"

	tele "gopkg.in/telebot.v3"

	"botmaker-backend/config"
	"botmaker-backend/internal/models"
	"botmaker-backend/internal/storage"
	"botmaker-backend/internal/templates"
)

// Sentinel errors returned by the engine.
var (
	ErrBotNotFound     = errors.New("engine: bot not found")
	ErrTemplateMissing = errors.New("engine: template not found")
)

// allowedUpdates restricts what Telegram delivers to the shared webhook.
var allowedUpdates = []string{
	"message",
	"edited_message",
	"channel_post",
	"edited_channel_post",
	"inline_query",
	"chosen_inline_result",
	"callback_query",
}

// Engine owns every live Telegram bot. All bots share a single webhook
// endpoint (<base>/api/webhook/<token>); incoming updates are routed to
// the matching bot via telebot's ProcessUpdate.
type Engine struct {
	store *storage.Store
	cfg   config.Config

	templates map[string]templates.Template

	mu   sync.RWMutex
	bots map[string]*tele.Bot // token -> live bot instance
}

// New creates an empty engine. Templates must be registered before use.
func New(store *storage.Store, cfg config.Config) *Engine {
	return &Engine{
		store:     store,
		cfg:       cfg,
		templates: make(map[string]templates.Template),
		bots:      make(map[string]*tele.Bot),
	}
}

// RegisterTemplate makes a template available to new bots.
func (e *Engine) RegisterTemplate(t templates.Template) {
	e.templates[t.Name()] = t
}

// HasTemplate reports whether a template with the given name exists.
func (e *Engine) HasTemplate(name string) bool {
	_, ok := e.templates[name]
	return ok
}

// TemplateInfos returns the metadata of every registered template.
func (e *Engine) TemplateInfos() []models.Template {
	names := make([]string, 0, len(e.templates))
	for name := range e.templates {
		names = append(names, name)
	}
	sort.Strings(names)

	infos := make([]models.Template, 0, len(names))
	for _, name := range names {
		t := e.templates[name]
		infos = append(infos, models.Template{
			Name:        t.Name(),
			Title:       t.Title(),
			Description: t.Description(),
			Commands:    t.Commands(),
		})
	}
	return infos
}

// Start activates every bot that is currently marked active.
// Bots that fail to come up are marked paused.
func (e *Engine) Start() error {
	bots, err := e.store.ListBots()
	if err != nil {
		return err
	}
	for i := range bots {
		b := &bots[i]
		if !b.IsActive() {
			continue
		}
		if err := e.activate(b); err != nil {
			log.Printf("[engine] activate bot %q failed: %v", b.Name, err)
			_ = e.store.SetBotStatus(b.ID, models.StatusPaused)
			continue
		}
		log.Printf("[engine] bot %q webhook active", b.Name)
	}
	return nil
}

// AddBot stores a new bot and brings it online. If Telegram rejects the
// token or the webhook, the bot is kept in the database but marked paused.
func (e *Engine) AddBot(b *models.Bot) error {
	if err := e.store.CreateBot(b); err != nil {
		return err
	}
	if err := e.activate(b); err != nil {
		b.Status = models.StatusPaused
		_ = e.store.SetBotStatus(b.ID, models.StatusPaused)
		return err
	}
	return nil
}

// ToggleBot flips the status of a bot between active and paused,
// registering or removing its webhook accordingly.
func (e *Engine) ToggleBot(id int64) (*models.Bot, error) {
	b, err := e.store.GetBot(id)
	if err != nil {
		return nil, err
	}
	if b.IsActive() {
		e.deactivate(b)
		if err := e.store.SetBotStatus(id, models.StatusPaused); err != nil {
			return nil, err
		}
	} else {
		if err := e.activate(b); err != nil {
			return nil, err
		}
		if err := e.store.SetBotStatus(id, models.StatusActive); err != nil {
			return nil, err
		}
	}
	return e.store.GetBot(id)
}

// DeleteBot removes the webhook, the live instance and the database record.
func (e *Engine) DeleteBot(id int64) error {
	b, err := e.store.GetBot(id)
	if err != nil {
		return err
	}
	e.deactivate(b)
	return e.store.DeleteBot(id)
}

// ProcessUpdate dispatches a raw Telegram update to the bot that owns
// the token. Returns ErrBotNotFound when the bot is unknown or paused.
func (e *Engine) ProcessUpdate(token string, body []byte) error {
	e.mu.RLock()
	bot, ok := e.bots[token]
	e.mu.RUnlock()
	if !ok {
		return ErrBotNotFound
	}

	var update tele.Update
	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("decode update: %w", err)
	}
	bot.ProcessUpdate(update)
	return nil
}

// SetWebhook registers a bot's webhook with Telegram so updates flow to
// the shared endpoint. The URL must be unique per bot.
func (e *Engine) SetWebhook(bot *tele.Bot, url string) error {
	return bot.SetWebhook(&tele.Webhook{
		Endpoint:       &tele.WebhookEndpoint{PublicURL: url},
		AllowedUpdates: allowedUpdates,
		DropUpdates:    true,
	})
}

// DeleteWebhook removes a bot's webhook from Telegram.
func (e *Engine) DeleteWebhook(bot *tele.Bot) error {
	return bot.RemoveWebhook(true)
}

// activate creates the live bot, applies its template and points the
// webhook at the shared endpoint.
func (e *Engine) activate(b *models.Bot) error {
	tmpl, ok := e.templates[b.Template]
	if !ok {
		return fmt.Errorf("%w: %q", ErrTemplateMissing, b.Template)
	}

	bot, err := tele.NewBot(tele.Settings{
		Token:     b.Token,
		ParseMode: tele.ModeHTML,
		OnError: func(err error, c tele.Context) {
			log.Printf("[bot %s] handler error: %v", b.Name, err)
		},
	})
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	if err := tmpl.Apply(bot, templates.Options{
		BotID:    b.ID,
		AdminIDs: e.cfg.AdminIDs,
		Store:    e.store,
	}); err != nil {
		return fmt.Errorf("apply template %q: %w", b.Template, err)
	}

	webhookURL := e.webhookURL(b.Token)
	if err := e.SetWebhook(bot, webhookURL); err != nil {
		return fmt.Errorf("set webhook: %w", err)
	}

	e.mu.Lock()
	e.bots[b.Token] = bot
	e.mu.Unlock()

	if err := e.store.SetBotWebhook(b.ID, webhookURL); err != nil {
		log.Printf("[engine] persist webhook url for bot %d: %v", b.ID, err)
	}
	return nil
}

// deactivate removes the live instance and unregisters the webhook.
func (e *Engine) deactivate(b *models.Bot) {
	e.mu.Lock()
	bot, ok := e.bots[b.Token]
	delete(e.bots, b.Token)
	e.mu.Unlock()

	if ok {
		if err := e.DeleteWebhook(bot); err != nil {
			log.Printf("[engine] delete webhook for bot %d: %v", b.ID, err)
		}
	}
	if err := e.store.SetBotWebhook(b.ID, ""); err != nil {
		log.Printf("[engine] clear webhook url for bot %d: %v", b.ID, err)
	}
}

func (e *Engine) webhookURL(token string) string {
	return fmt.Sprintf("%s/api/webhook/%s", e.cfg.WebhookBaseURL, token)
}
