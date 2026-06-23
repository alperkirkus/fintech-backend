package handler

import (
	"encoding/json"
	"net/http"

	"github.com/alperkirkus/fintech-backend/internal/middleware"
	"github.com/alperkirkus/fintech-backend/internal/service"
)

type AuthHandler struct {
	auth  service.AuthService
	users service.UserService
}

func NewAuthHandler(auth service.AuthService, users service.UserService) *AuthHandler {
	return &AuthHandler{auth: auth, users: users}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.users.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	token, err := h.auth.Refresh(claims)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not refresh token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
