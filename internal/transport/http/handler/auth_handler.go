package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	service "go-api-kbt/internal/service/auth"
	"go-api-kbt/internal/transport/http/dto"
	httputil "go-api-kbt/internal/transport/httputil"

	"github.com/go-chi/chi/v5"
)

// AuthHandler exposes authentication related HTTP handlers.
type AuthHandler struct {
	service *service.Service
	logger  *slog.Logger
}

// NewAuthHandler constructs a new handler.
func NewAuthHandler(service *service.Service, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{service: service, logger: logger}
}

// RegisterRoutes attaches the handler routes to the router group.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.login)
	r.Post("/auth/refresh", h.refresh)
	r.Post("/auth/logout", h.logout)
}

// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.LoginInput true "Login credentials"
// @Success 200 {object} dto.AuthOutput
// @Failure 400 {object} dto.Problem
// @Failure 401 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /auth/login [post]
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	output, err := h.service.Login(r.Context(), input)
	if err != nil {
		if httputil.RespondValidationWithJSONTags(w, input, err) {
			return
		}
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			httputil.RespondError(w, http.StatusUnauthorized, err.Error())
		default:
			h.logger.Error("login failed", slog.String("error", err.Error()))
			httputil.RespondError(w, http.StatusInternalServerError, "failed to login")
		}
		return
	}

	httputil.RespondJSON(w, http.StatusOK, output)
}

// @Summary Refresh JWT tokens
// @Description Generate new access and refresh tokens from a valid refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.RefreshTokenInput true "Refresh token"
// @Success 200 {object} dto.AuthOutput
// @Failure 400 {object} dto.Problem
// @Failure 401 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /auth/refresh [post]
func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var input dto.RefreshTokenInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	output, err := h.service.Refresh(r.Context(), input.RefreshToken)
	if err != nil {
		if httputil.TryRespondValidation(w, err) {
			return
		}
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			httputil.RespondError(w, http.StatusUnauthorized, err.Error())
		default:
			h.logger.Error("refresh token failed", slog.String("error", err.Error()))
			httputil.RespondError(w, http.StatusInternalServerError, "failed to refresh token")
		}
		return
	}

	httputil.RespondJSON(w, http.StatusOK, output)
}

// @Summary Logout user
// @Description Invalidate a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param input body dto.RefreshTokenInput true "Refresh token"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /auth/logout [post]
func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	var input dto.RefreshTokenInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := h.service.Logout(r.Context(), input.RefreshToken); err != nil {
		if httputil.TryRespondValidation(w, err) {
			return
		}
		h.logger.Error("logout failed", slog.String("error", err.Error()))
		httputil.RespondError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
