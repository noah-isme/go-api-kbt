package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	repo "go-api-kbt/internal/repository/user"
	service "go-api-kbt/internal/service/user"

	"github.com/go-chi/chi/v5"
)

// UserHandler exposes user related HTTP handlers.
type UserHandler struct {
	service *service.Service
	logger  *slog.Logger
}

// NewUserHandler constructs a new handler.
func NewUserHandler(service *service.Service, logger *slog.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

// RegisterRoutes attaches the handler routes to the router group.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Post("/users", h.create)
	r.Get("/users", h.list)
	r.Route("/users/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Create(r.Context(), input)
	if err != nil {
		// validation errors
		if respondValidationWithJSONTags(w, input, err) {
			return
		}

		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			respondError(w, http.StatusConflict, err.Error())
		default:
			h.logger.Error("create user", slog.String("error", err.Error()))
			respondError(w, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	respondJSON(w, http.StatusCreated, dto)
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	users, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("list users", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data":   users,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	dto, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("get user", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Read body for better debug logging on decode errors.
	body, _ := io.ReadAll(r.Body)
	var input service.UpdateInput
	if err := json.Unmarshal(body, &input); err != nil {
		h.logger.Error("failed to decode update payload", slog.String("error", err.Error()), slog.String("body", string(body)))
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if tryRespondValidation(w, err) {
			return
		}
		if errors.Is(err, repo.ErrNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("update user", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("delete user", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parsePagination(r *http.Request) (int, int) {
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// response writing errors are logged but not returned to the client
		slog.Default().Error("encode response", slog.String("error", err.Error()))
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
