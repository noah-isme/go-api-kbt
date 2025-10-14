package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	repo "go-api-kbt/internal/repository/location"
	service "go-api-kbt/internal/service/location"
	web "go-api-kbt/internal/transport/http"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-chi/chi/v5"
)

// LocationHandler exposes location related HTTP handlers.
type LocationHandler struct {
	service *service.Service
	logger  *slog.Logger
}

// NewLocationHandler constructs a new handler.
func NewLocationHandler(service *service.Service, logger *slog.Logger) *LocationHandler {
	return &LocationHandler{service: service, logger: logger}
}

// RegisterRoutes attaches the handler routes to the router group.
func (h *LocationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/locations", h.create)
	r.Get("/locations", h.list)
	r.Route("/locations/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})
}

// @Summary Create a new location
// @Description Create a new location with the provided details
// @Tags Location
// @Accept json
// @Produce json
// @Param input body dto.CreateLocationInput true "Location creation input"
// @Success 201 {object} dto.LocationDTO
// @Failure 400 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /locations [post]
func (h *LocationHandler) create(w http.ResponseWriter, r *http.Request) {
	var input service.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Create(r.Context(), input)
	if err != nil {
		if RespondValidationWithJSONTags(w, input, err) {
			return
		}
		h.logger.Error("create location", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to create location")
		return
	}

	respondJSON(w, http.StatusCreated, dto)
}

// @Summary Get all locations
// @Description Retrieve a list of all locations with pagination and filtering
// @Tags Location
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Param q query string false "Search query"
// @Param sort query string false "Sort order (e.g., id asc, name desc)"
// @Success 200 {object} dto.LocationListResponse
// @Failure 500 {object} dto.Problem
// @Router /locations [get]
func (h *LocationHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, offset := web.ParsePagination(r)
	locations, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("list locations", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list locations")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data":   locations,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary Get a location by ID
// @Description Retrieve a single location by its ID
// @Tags Location
// @Produce json
// @Param id path int true "Location ID"
// @Success 200 {object} dto.LocationDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /locations/{id} [get]
func (h *LocationHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid location id")
		return
	}

	dto, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			respondError(w, http.StatusNotFound, "location not found")
			return
		}
		h.logger.Error("get location", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get location")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}

// @Summary Update a location
// @Description Update an existing location with the provided details
// @Tags Location
// @Accept json
// @Produce json
// @Param id path int true "Location ID"
// @Param input body dto.UpdateLocationInput true "Location update input"
// @Success 200 {object} dto.LocationDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /locations/{id} [put]
func (h *LocationHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid location id")
		return
	}

	var input service.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if TryRespondValidation(w, err) {
			return
		}
		if errors.Is(err, repo.ErrNotFound) {
			respondError(w, http.StatusNotFound, "location not found")
			return
		}
		h.logger.Error("update location", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to update location")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}

// @Summary Delete a location
// @Description Delete a location by its ID
// @Tags Location
// @Produce json
// @Param id path int true "Location ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /locations/{id} [delete]
func (h *LocationHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid location id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			respondError(w, http.StatusNotFound, "location not found")
			return
		}
		h.logger.Error("delete location", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to delete location")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


