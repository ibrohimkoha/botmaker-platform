package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"botmaker-backend/internal/models"
)

// handleListCustomTemplates returns the admin-submitted repo templates.
func (a *API) handleListCustomTemplates(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	templates, err := a.store.ListCustomRepoTemplates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Shablonlarni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

// handleCreateCustomTemplate registers a new git-repo-backed template.
func (a *API) handleCreateCustomTemplate(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		GitRepoURL  string `json:"git_repo_url"`
		Category    string `json:"category"`
		Price       int64  `json:"price"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Title = strings.TrimSpace(req.Title)
	req.GitRepoURL = strings.TrimSpace(req.GitRepoURL)
	req.Category = strings.TrimSpace(req.Category)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "shablon nomi (name) bo'sh bo'lishi mumkin emas")
		return
	}
	if req.Title == "" {
		req.Title = req.Name
	}
	if req.GitRepoURL == "" {
		writeError(w, http.StatusBadRequest, "git_repo_url bo'sh bo'lishi mumkin emas")
		return
	}
	if req.Price < 0 {
		writeError(w, http.StatusBadRequest, "narx manfiy bo'lishi mumkin emas")
		return
	}

	t := &models.CustomRepoTemplate{
		Name:        req.Name,
		Title:       req.Title,
		Description: strings.TrimSpace(req.Description),
		GitRepoURL:  req.GitRepoURL,
		Category:    req.Category,
		Price:       req.Price,
	}
	if err := a.store.CreateCustomRepoTemplate(t); err != nil {
		writeError(w, http.StatusInternalServerError, "Shablon qo'shib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

// handleListUsers returns all registered users.
func (a *API) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.store.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Foydalanuvchilarni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, users)
}
