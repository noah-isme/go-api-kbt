package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	repo "go-api-kbt/internal/repository/user"
	service "go-api-kbt/internal/service/user"
	web "go-api-kbt/internal/transport/http"
	"go-api-kbt/internal/transport/http/dto"

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

// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags User
// @Accept json
// @Produce json
// @Param input body dto.CreateUserInput true "User creation input"
// @Success 201 {object} dto.UserDTO
// @Failure 400 {object} dto.Problem
// @Failure 409 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /users [post]
func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Create(r.Context(), input)
	if err != nil {
		// validation errors
		if RespondValidationWithJSONTags(w, input, err) {
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

// @Summary Get all users
// @Description Retrieve a list of all users with pagination and filtering
// @Tags User
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Param q query string false "Search query"
// @Param sort query string false "Sort order (e.g., id asc, name desc)"
// @Success 200 {object} dto.UserListResponse
// @Failure 500 {object} dto.Problem
// @Router /users [get]
func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, offset := web.ParsePagination(r)
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

// @Summary Get a user by ID
// @Description Retrieve a single user by its ID
// @Tags User
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /users/{id} [get]
func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
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

// @Summary Update a user
// @Description Update an existing user with the provided details
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param input body dto.UpdateUserInput true "User update input"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /users/{id} [put]
func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Read body for better debug logging on decode errors.
	body, _ := io.ReadAll(r.Body)
	var input dto.UpdateUserInput
	if err := json.Unmarshal(body, &input); err != nil {
		h.logger.Error("failed to decode update payload", slog.String("error", err.Error()), slog.String("body", string(body)))
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if TryRespondValidation(w, err) {
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

// @Summary Delete a user
// @Description Delete a user by its ID
// @Tags User
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /users/{id} [delete]
func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
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
