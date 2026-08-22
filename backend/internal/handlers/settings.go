package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"botmaker-backend/internal/storage"
)

// handleGetBotSettings returns every stored setting of a bot together
// with its webhook URL.
func (a *API) handleGetBotSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireUser(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	bot, err := a.store.GetBot(id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "bot topilmadi")
			return
		}
		writeError(w, http.StatusInternalServerError, "Botni o'qib bo'lmadi")
		return
	}
	settings, err := a.store.GetAllSettings(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Sozlamalarni o'qib bo'lmadi")
		return
	}
	if settings == nil {
		settings = map[string]string{}
	}
	settings["webhook_url"] = bot.WebhookURL
	writeJSON(w, http.StatusOK, map[string]any{
		"id":       bot.ID,
		"name":     bot.Name,
		"username": bot.Username,
		"status":   bot.Status,
		"settings": settings,
	})
}

// handleSaveBotSettings upserts the submitted settings of a bot.
func (a *API) handleSaveBotSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri bot id")
		return
	}
	if _, err := a.store.GetBot(id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "bot topilmadi")
			return
		}
		writeError(w, http.StatusInternalServerError, "Botni o'qib bo'lmadi")
		return
	}
	var settings map[string]string
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	if webhook, ok := settings["webhook_url"]; ok {
		if err := a.store.SetBotWebhook(id, webhook); err != nil {
			writeError(w, http.StatusInternalServerError, "Webhook saqlanmadi")
			return
		}
		delete(settings, "webhook_url")
	}
	if err := a.store.SetAllSettings(id, settings); err != nil {
		writeError(w, http.StatusInternalServerError, "Sozlamalarni saqlab bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
