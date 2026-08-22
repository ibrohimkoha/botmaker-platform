package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"botmaker-backend/internal/engine"
	"botmaker-backend/internal/models"
	"botmaker-backend/internal/storage"
)

// API exposes the REST management interface of the platform.
type API struct {
	engine *engine.Engine
	store  *storage.Store
}

// NewAPI builds the handler set.
func NewAPI(e *engine.Engine, s *storage.Store) *API {
	return &API{engine: e, store: s}
}

// Routes returns the fully wired HTTP handler with CORS enabled.
func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("GET /api/bots", a.handleListBots)
	mux.HandleFunc("POST /api/bots", a.handleCreateBot)
	mux.HandleFunc("DELETE /api/bots/{id}", a.handleDeleteBot)
	mux.HandleFunc("POST /api/bots/{id}/toggle", a.handleToggleBot)
	mux.HandleFunc("GET /api/templates", a.handleListTemplates)
	mux.HandleFunc("GET /api/stats", a.handleStats)
	mux.HandleFunc("POST /api/webhook/{token}", a.handleWebhook)
	mux.HandleFunc("POST /webhook/{token}", a.handleWebhook)
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
