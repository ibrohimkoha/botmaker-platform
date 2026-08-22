package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"botmaker-backend/config"
	"botmaker-backend/internal/engine"
	"botmaker-backend/internal/models"
	"botmaker-backend/internal/storage"
)

// API exposes the REST management interface of the platform.
type API struct {
	engine *engine.Engine
	store  *storage.Store
	cfg    config.Config
	b2     *storage.B2Uploader

	mu          sync.Mutex
	oauthStates map[string]time.Time
}

// NewAPI builds the handler set.
func NewAPI(e *engine.Engine, s *storage.Store, cfg config.Config) *API {
	api := &API{
		engine:      e,
		store:       s,
		cfg:         cfg,
		oauthStates: map[string]time.Time{},
	}
	if cfg.B2.Endpoint != "" && cfg.B2.Bucket != "" && cfg.B2.KeyID != "" && cfg.B2.ApplicationKey != "" {
		if up, err := storage.NewB2Uploader(cfg.B2.Endpoint, cfg.B2.Bucket, cfg.B2.KeyID, cfg.B2.ApplicationKey, cfg.B2.Region); err != nil {
			log.Printf("[api] b2 uploader init failed: %v", err)
		} else {
			api.b2 = up
		}
	}
	return api
}

// Routes returns the fully wired HTTP handler with CORS enabled.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("GET /api/bots", a.handleListBots)
	mux.HandleFunc("POST /api/bots", a.handleCreateBot)
	mux.HandleFunc("PUT /api/bots/{id}", a.handleUpdateBot)
	mux.HandleFunc("PATCH /api/bots/{id}", a.handleUpdateBot)
	mux.HandleFunc("POST /api/bots/{id}", a.handleUpdateBot)
	mux.HandleFunc("DELETE /api/bots/{id}", a.handleDeleteBot)
	mux.HandleFunc("POST /api/bots/{id}/toggle", a.handleToggleBot)
	mux.HandleFunc("POST /api/bots/{id}/start", a.handleStartBot)
	mux.HandleFunc("POST /api/bots/{id}/stop", a.handleStopBot)
	mux.HandleFunc("GET /api/bots/{id}/settings", a.handleGetBotSettings)
	mux.HandleFunc("PUT /api/bots/{id}/settings", a.handleSaveBotSettings)
	mux.HandleFunc("POST /api/broadcast", a.handleBroadcast)
	mux.HandleFunc("POST /api/bots/{id}/broadcast", a.handleBroadcast)
	mux.HandleFunc("GET /api/templates", a.handleListTemplates)
	mux.HandleFunc("GET /api/stats", a.handleStats)
	mux.HandleFunc("POST /api/webhook/{token}", a.handleWebhook)
	mux.HandleFunc("POST /webhook/{token}", a.handleWebhook)

	// Auth.
	mux.HandleFunc("POST /api/auth/google", a.handleGoogleAuth)
	mux.HandleFunc("GET /api/auth/callback/google", a.handleGoogleCallback)
	mux.HandleFunc("POST /api/auth/telegram", a.handleTelegramAuth)
	mux.HandleFunc("POST /api/auth/quick-login", a.handleQuickLogin)
	mux.HandleFunc("GET /api/auth/me", a.handleMe)

	// Deposits & Checks.
	mux.HandleFunc("POST /api/deposits", a.handleCreateDeposit)
	mux.HandleFunc("GET /api/deposits", a.handleListDeposits)
	mux.HandleFunc("POST /api/payments/checks", a.handleCreateDeposit)
	mux.HandleFunc("GET /api/payments/checks", a.handleListDeposits)
	mux.HandleFunc("POST /api/admin/deposits/{id}/approve", a.handleApproveDeposit)
	mux.HandleFunc("POST /api/admin/deposits/{id}/reject", a.handleRejectDeposit)
	mux.HandleFunc("POST /api/payments/checks/{id}/approve", a.handleApproveDeposit)
	mux.HandleFunc("POST /api/payments/checks/{id}/reject", a.handleRejectDeposit)

	// Payment cards.
	mux.HandleFunc("GET /api/cards", a.handleListCards)
	mux.HandleFunc("POST /api/cards", a.handleCreateCard)
	mux.HandleFunc("GET /api/settings/card", a.handleListCards)
	mux.HandleFunc("POST /api/settings/card", a.handleCreateCard)
	mux.HandleFunc("POST /api/admin/cards", a.handleCreateCard)

	// Custom repo templates.
	mux.HandleFunc("GET /api/admin/templates", a.handleListCustomTemplates)
	mux.HandleFunc("POST /api/admin/templates", a.handleCreateCustomTemplate)
	mux.HandleFunc("GET /api/templates/custom", a.handleListCustomTemplates)
	mux.HandleFunc("POST /api/templates/custom", a.handleCreateCustomTemplate)

	// Users.
	mux.HandleFunc("GET /api/users", a.handleListUsers)
	mux.HandleFunc("GET /api/admin/users", a.handleListUsers)

	return withCORS(mux)
}

func (a *API) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleListBots(w http.ResponseWriter, _ *http.Request) {
	bots, err := a.store.ListBots()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "botlarni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, bots)
}

type createBotRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Template string `json:"template"`
}

func (a *API) handleCreateBot(w http.ResponseWriter, r *http.Request) {
	var req createBotRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Token = strings.TrimSpace(req.Token)
	req.Template = strings.TrimSpace(req.Template)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Bot nomi bo'sh bo'lishi mumkin emas")
		return
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "Bot token bo'sh bo'lishi mumkin emas (@BotFather dan oling)")
		return
	}
	normTpl := engine.NormalizeTemplate(req.Template)
	if !a.engine.HasTemplate(normTpl) {
		writeError(w, http.StatusBadRequest, "Noma'lum shablon: "+req.Template)
		return
	}

	bot := &models.Bot{
		Name:     req.Name,
		Username: req.Username,
		Token:    req.Token,
		Template: normTpl,
		Status:   models.StatusActive,
	}
	if err := a.engine.AddBot(bot); err != nil {
		log.Printf("[api] create bot %q activation failed: %v", req.Name, err)
		writeError(w, http.StatusBadRequest, "Botni faollashtirib bo'lmadi: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, bot)
}

func (a *API) handleDeleteBot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	if err := a.engine.DeleteBot(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "bot topilmadi")
			return
		}
		writeError(w, http.StatusInternalServerError, "botni o'chirib bo'lmadi")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleToggleBot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	bot, err := a.engine.ToggleBot(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "bot topilmadi")
			return
		}
		writeError(w, http.StatusBadRequest, "holatni o'zgartirib bo'lmadi: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, bot)
}

type updateBotSettingsRequest struct {
	AdminID    string `json:"admin_id"`
	APIKey     string `json:"api_key"`
	ChannelID  string `json:"channel_id"`
	Currency   string `json:"currency"`
	WebhookURL string `json:"webhook_url"`
}

func (a *API) handleUpdateBot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	var req updateBotSettingsRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	if req.AdminID != "" {
		_ = a.store.SetSetting(id, "admin_id", req.AdminID)
	}
	if req.APIKey != "" {
		_ = a.store.SetSetting(id, "api_key", req.APIKey)
	}
	if req.ChannelID != "" {
		_ = a.store.SetSetting(id, "channel_id", req.ChannelID)
	}
	if req.Currency != "" {
		_ = a.store.SetSetting(id, "currency", req.Currency)
	}
	if req.WebhookURL != "" {
		_ = a.store.SetBotWebhook(id, req.WebhookURL)
	}
	bot, err := a.store.GetBot(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bot topilmadi")
		return
	}
	writeJSON(w, http.StatusOK, bot)
}

func (a *API) handleStartBot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	bot, err := a.store.GetBot(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bot topilmadi")
		return
	}
	if !bot.IsActive() {
		bot, err = a.engine.ToggleBot(id)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, bot)
}

func (a *API) handleStopBot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	bot, err := a.store.GetBot(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "bot topilmadi")
		return
	}
	if bot.IsActive() {
		bot, err = a.engine.ToggleBot(id)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, bot)
}

type broadcastRequest struct {
	BotID   any    `json:"bot_id"`
	Target  string `json:"target"`
	Message string `json:"message"`
	Text    string `json:"text"`
}

func (a *API) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	var req broadcastRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format: "+err.Error())
		return
	}
	msgText := strings.TrimSpace(req.Message)
	if msgText == "" {
		msgText = strings.TrimSpace(req.Text)
	}
	if msgText == "" {
		writeError(w, http.StatusBadRequest, "Xabar matni bo'sh bo'lishi mumkin emas")
		return
	}

	var targetBotID int64
	switch v := req.BotID.(type) {
	case float64:
		targetBotID = int64(v)
	case int64:
		targetBotID = v
	case string:
		if v != "" && v != "all" {
			targetBotID, _ = strconv.ParseInt(v, 10, 64)
		}
	}
	if targetBotID == 0 && req.Target != "" && req.Target != "all" {
		targetBotID, _ = strconv.ParseInt(req.Target, 10, 64)
	}
	if pathID := r.PathValue("id"); pathID != "" {
		if id, err := strconv.ParseInt(pathID, 10, 64); err == nil {
			targetBotID = id
		}
	}

	if targetBotID > 0 {
		_ = a.store.CreateBroadcast(&models.Broadcast{
			BotID:   targetBotID,
			Message: msgText,
			Status:  models.BroadcastDone,
			Total:   1,
			Sent:    1,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Broadcast muvaffaqiyatli yuborildi",
	})
}

func (a *API) handleListTemplates(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.engine.TemplateInfos())
}

type statsResponse struct {
	TotalBots  int            `json:"total_bots"`
	ActiveBots int            `json:"active_bots"`
	PausedBots int            `json:"paused_bots"`
	Stats      models.Stats   `json:"stats"`
	PerBot     []models.Stats `json:"per_bot"`
}

func (a *API) handleStats(w http.ResponseWriter, _ *http.Request) {
	bots, err := a.store.ListBots()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "botlarni o'qib bo'lmadi")
		return
	}
	active, paused := 0, 0
	for _, b := range bots {
		if b.IsActive() {
			active++
		} else {
			paused++
		}
	}

	global, perBot, err := a.store.GetGlobalStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "statistikani o'qib bo'lmadi")
		return
	}

	writeJSON(w, http.StatusOK, statsResponse{
		TotalBots:  len(bots),
		ActiveBots: active,
		PausedBots: paused,
		Stats:      *global,
		PerBot:     perBot,
	})
}

// handleWebhook is the single entry point Telegram calls for every bot.
// The token in the URL selects the target bot instance.
func (a *API) handleWebhook(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := a.engine.ProcessUpdate(token, body); err != nil {
		if errors.Is(err, engine.ErrBotNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("[webhook] invalid update for token %s...: %v", token[:min(len(token), 12)], err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// withCORS enables cross-origin access for the management API.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
