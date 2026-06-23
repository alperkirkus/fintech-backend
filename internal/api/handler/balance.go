package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alperkirkus/fintech-backend/internal/middleware"
	"github.com/alperkirkus/fintech-backend/internal/service"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type BalanceHandler struct {
	balances service.BalanceService
}

func NewBalanceHandler(balances service.BalanceService) *BalanceHandler {
	return &BalanceHandler{balances: balances}
}

func (h *BalanceHandler) Current(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := parseClaimsUserID(w, claims.UserID)
	if !ok {
		return
	}

	balance, err := h.balances.GetBalance(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{
				"user_id": userID,
				"amount":  "0",
			})
			return
		}
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, balance)
}

func (h *BalanceHandler) Historical(w http.ResponseWriter, r *http.Request) {
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

	transactionHistory, err := h.balances.GetHistory(r.Context(), userID, limit, offset)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, transactionHistory)
}

func (h *BalanceHandler) AtTime(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, ok := parseClaimsUserID(w, claims.UserID)
	if !ok {
		return
	}

	atStr := r.URL.Query().Get("at")
	if atStr == "" {
		writeError(w, http.StatusBadRequest, "missing 'at' query param (RFC3339)")
		return
	}

	at, err := time.Parse(time.RFC3339, atStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'at' format, use RFC3339")
		return
	}

	amount, err := h.balances.GetAtTime(r.Context(), userID, at)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": userID,
		"amount":  amount,
		"at":      at,
	})
}
