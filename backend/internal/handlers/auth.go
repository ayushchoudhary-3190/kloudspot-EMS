package handlers

import (
	"net/http"

	"EMS/internal/models"
	"EMS/internal/service"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	svc *service.AuthService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login authenticates an admin and returns a JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in models.LoginInput
	if err := decodeJSON(r, &in); err != nil {
		badRequest(w, "invalid request body")
		return
	}

	token, err := h.svc.Login(in)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token":    token,
		"username": in.Username,
	})
}
