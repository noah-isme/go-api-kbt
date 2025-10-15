package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	repo "go-api-kbt/internal/repository/event"
	service "go-api-kbt/internal/service/event"
	"go-api-kbt/internal/transport/http/dto"
	httputil "go-api-kbt/internal/transport/httputil"

	"github.com/go-chi/chi/v5"
)

// EventHandler exposes event related HTTP handlers.
type EventHandler struct {
	service *service.Service
	logger  *slog.Logger
}

// NewEventHandler constructs a new handler.
func NewEventHandler(service *service.Service, logger *slog.Logger) *EventHandler {
	return &EventHandler{service: service, logger: logger}
}

// RegisterRoutes attaches the handler routes to the router group.
func (h *EventHandler) RegisterRoutes(r chi.Router) {
	r.Post("/events", h.create)
	r.Get("/events", h.list)
	r.Route("/events/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})
}

// @Summary Create a new event
// @Description Create a new event with the provided details
// @Tags Event
// @Accept json
// @Produce json
// @Param input body dto.CreateEventInput true "Event creation input"
// @Success 201 {object} dto.EventDTO
// @Failure 400 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /events [post]
func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Create(r.Context(), input)
	if err != nil {
		if httputil.RespondValidationWithJSONTags(w, input, err) {
			return
		}
		h.logger.Error("create event", slog.String("error", err.Error()))
		httputil.RespondError(w, http.StatusInternalServerError, "failed to create event")
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, dto)
}

// @Summary Get all events
// @Description Retrieve a list of all events with pagination and filtering
// @Tags Event
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Param q query string false "Search query"
// @Param sort query string false "Sort order (e.g., id asc, name desc)"
// @Success 200 {object} dto.EventListResponse
// @Failure 500 {object} dto.Problem
// @Router /events [get]
func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
	params := httputil.ParsePaginationParams(r)
	response, err := h.service.List(r.Context(), params)
	if err != nil {
		h.logger.Error("list events", slog.String("error", err.Error()))
		httputil.RespondError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	httputil.RespondJSON(w, http.StatusOK, response)
}

// @Summary Get an event by ID
// @Description Retrieve a single event by its ID
// @Tags Event
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} dto.EventDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /events/{id} [get]
func (h *EventHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	dto, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		h.logger.Error("get event", slog.String("error", err.Error()))
		httputil.RespondError(w, http.StatusInternalServerError, "failed to get event")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, dto)
}

// @Summary Update an event
// @Description Update an existing event with the provided details
// @Tags Event
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param input body dto.UpdateEventInput true "Event update input"
// @Success 200 {object} dto.EventDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /events/{id} [put]
func (h *EventHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var input dto.UpdateEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if httputil.TryRespondValidation(w, err) {
			return
		}
		if errors.Is(err, repo.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		h.logger.Error("update event", slog.String("error", err.Error()))
		httputil.RespondError(w, http.StatusInternalServerError, "failed to update event")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, dto)
}

// @Summary Delete an event
// @Description Delete an event by its ID
// @Tags Event
// @Produce json
// @Param id path int true "Event ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /events/{id} [delete]
func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			httputil.RespondError(w, http.StatusNotFound, "event not found")
			return
		}
		h.logger.Error("delete event", slog.String("error", err.Error()))
		httputil.RespondError(w, http.StatusInternalServerError, "failed to delete event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
