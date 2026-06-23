package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/alperkirkus/fintech-backend/internal/middleware"
	"github.com/alperkirkus/fintech-backend/internal/service"
)

type TransactionHandler struct {
	transactions   service.TransactionService
	balanceService service.BalanceService
}

func NewTransactionHandler(transactions service.TransactionService, balanceService service.BalanceService) *TransactionHandler {
	return &TransactionHandler{transactions: transactions, balanceService: balanceService}
}

func (h *TransactionHandler) Credit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := parseClaimsUserID(w, claims.UserID)
	if !ok {
		return
	}

	var req struct {
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	transaction, err := h.transactions.Deposit(r.Context(), userID, amount)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transaction)
}

func (h *TransactionHandler) Debit(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := parseClaimsUserID(w, claims.UserID)
	if !ok {
		return
	}

	var req struct {
		Amount string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	transaction, err := h.transactions.Withdraw(r.Context(), userID, amount)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transaction)
}

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	fromUserID, ok := parseClaimsUserID(w, claims.UserID)
	if !ok {
		return
	}

	var req struct {
		ToUserID string `json:"to_user_id"`
		Amount   string `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to_user_id")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	transaction, err := h.transactions.Transfer(r.Context(), fromUserID, toUserID, amount)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, transaction)
}

func (h *TransactionHandler) History(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := parseClaimsUserID(w, claims.UserID)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	transactionHistory, err := h.balanceService.GetHistory(r.Context(), userID, limit, offset)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transactionHistory)
}

func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid transaction id")
		return
	}

	_, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	transaction, err := h.transactions.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transaction)
}
