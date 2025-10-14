package handler

import (
	"log/slog"
	"net/http"

	service "go-api-kbt/internal/service/activity"
	web "go-api-kbt/internal/transport/http"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-chi/chi/v5"
)

// ActivityHandler exposes activity related HTTP handlers.
type ActivityHandler struct {
	service *service.Service
	logger  *slog.Logger
}

// NewActivityHandler constructs a new handler.
func NewActivityHandler(service *service.Service, logger *slog.Logger) *ActivityHandler {
	return &ActivityHandler{service: service, logger: logger}
}

// RegisterRoutes attaches the handler routes to the router group.
func (h *ActivityHandler) RegisterRoutes(r chi.Router) {
	r.Get("/activities", h.list)
	r.Get("/activities/{id}", h.get)
}

// @Summary Get all activities
// @Description Retrieve a list of all activities with pagination and filtering
// @Tags Activity
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Param q query string false "Search query"
// @Param sort query string false "Sort order (e.g., id asc, name desc)"
// @Success 200 {object} dto.ActivityListResponse
// @Failure 500 {object} dto.Problem
// @Router /activities [get]
func (h *ActivityHandler) list(w http.ResponseWriter, r *http.Request) {
	params := web.ParsePaginationParams(r)

	input := service.ListInput{
		Page:  params.Page,
		Limit: params.Limit,
		Query: params.Query,
		Sort:  params.Sort,
	}

	activities, meta, err := h.service.List(r.Context(), input)
	if err != nil {
		h.logger.Error("list activities", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list activities")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data": activities,
		"meta": meta,
	})
}

// @Summary Get an activity by ID
// @Description Retrieve a single activity by its ID
// @Tags Activity
// @Produce json
// @Param id path int true "Activity ID"
// @Success 200 {object} dto.ActivityDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /activities/{id} [get]
func (h *ActivityHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid activity id")
		return
	}

	dto, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		// Using a generic error for now, in a real app you'd check for specific errors like ErrNotFound
		respondError(w, http.StatusNotFound, "activity not found")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}
