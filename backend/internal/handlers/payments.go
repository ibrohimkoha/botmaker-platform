package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"botmaker-backend/internal/models"
	"botmaker-backend/internal/storage"
)

// handleCreateDeposit accepts a multipart form with a payment receipt
// (field "receipt") and an amount, uploads the receipt to B2 and
// creates a pending deposit.
func (a *API) handleCreateDeposit(w http.ResponseWriter, r *http.Request) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	if a.b2 == nil {
		writeError(w, http.StatusServiceUnavailable, "Chek yuklash hizmati sozlanmagan")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "multipart forma noto'g'ri: "+err.Error())
		return
	}

	amount, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("amount")), 10, 64)
	if err != nil || amount <= 0 {
		writeError(w, http.StatusBadRequest, "to'lov summasi noto'g'ri")
		return
	}

	file, header, err := r.FormFile("receipt")
	if err != nil {
		writeError(w, http.StatusBadRequest, "chek fayli (receipt) yuklanmadi")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("receipts/%d/%d-%s%s", u.ID, time.Now().Unix(), randHex(8), ext)
	receiptURL, err := a.b2.UploadMultipartFile(r.Context(), objectName, file, header)
	if err != nil {
		log.Printf("[payments] receipt upload failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Chekni yuklab bo'lmadi")
		return
	}

	d := &models.Deposit{
		UserID:       u.ID,
		UserFullName: u.FullName,
		Amount:       amount,
		ReceiptURL:   receiptURL,
		Status:       models.DepositPending,
	}
	if err := a.store.CreateDeposit(d); err != nil {
		log.Printf("[payments] create deposit: %v", err)
		writeError(w, http.StatusInternalServerError, "Depozit yaratib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

// handleListDeposits returns all deposits for admins and only the
// caller's deposits for regular users.
func (a *API) handleListDeposits(w http.ResponseWriter, r *http.Request) {
	u, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	var (
		list []models.Deposit
		err  error
	)
	if a.isAdmin(u) {
		list, err = a.store.ListDeposits()
	} else {
		list, err = a.store.ListDepositsByUser(u.ID)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Depozitlarni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// handleApproveDeposit approves a pending deposit and credits the
// user's balance.
func (a *API) handleApproveDeposit(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri depozit id")
		return
	}
	if err := a.store.ApproveDeposit(id); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			writeError(w, http.StatusNotFound, "depozit topilmadi")
		case errors.Is(err, storage.ErrDepositProcessed):
			writeError(w, http.StatusConflict, "depozit allaqachon ko'rib chiqilgan")
		default:
			writeError(w, http.StatusInternalServerError, "Depozitni tasdiqlab bo'lmadi")
		}
		return
	}
	d, err := a.store.GetDeposit(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Depozitni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// handleRejectDeposit rejects a pending deposit with a reason.
func (a *API) handleRejectDeposit(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri depozit id")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req)
	reason := strings.TrimSpace(req.Reason)

	if err := a.store.RejectDeposit(id, reason); err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			writeError(w, http.StatusNotFound, "depozit topilmadi")
		case errors.Is(err, storage.ErrDepositProcessed):
			writeError(w, http.StatusConflict, "depozit allaqachon ko'rib chiqilgan")
		default:
			writeError(w, http.StatusInternalServerError, "Depozitni rad etib bo'lmadi")
		}
		return
	}
	d, err := a.store.GetDeposit(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Depozitni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// handleListCards returns the active payment cards for transfers.
func (a *API) handleListCards(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireUser(w, r); !ok {
		return
	}
	cards, err := a.store.ListActiveCards()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Kartalarni o'qib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusOK, cards)
}

// handleCreateCard registers a new payment card (admin only).
func (a *API) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.requireAdmin(w, r); !ok {
		return
	}
	var req struct {
		CardNumber string `json:"card_number"`
		CardHolder string `json:"card_holder"`
		BankName   string `json:"bank_name"`
		IsActive   *bool  `json:"is_active"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "noto'g'ri JSON format")
		return
	}
	req.CardNumber = strings.TrimSpace(req.CardNumber)
	req.CardHolder = strings.TrimSpace(req.CardHolder)
	req.BankName = strings.TrimSpace(req.BankName)
	if req.CardNumber == "" {
		writeError(w, http.StatusBadRequest, "karta raqami bo'sh bo'lishi mumkin emas")
		return
	}

	card := &models.PaymentCard{
		CardNumber: req.CardNumber,
		CardHolder: req.CardHolder,
		BankName:   req.BankName,
		IsActive:   true,
	}
	if req.IsActive != nil {
		card.IsActive = *req.IsActive
	}
	if err := a.store.CreateCard(card); err != nil {
		writeError(w, http.StatusInternalServerError, "Karta qo'shib bo'lmadi")
		return
	}
	writeJSON(w, http.StatusCreated, card)
}
