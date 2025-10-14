package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	service "go-api-kbt/internal/service/medaler"
	web "go-api-kbt/internal/transport/http"
	"go-api-kbt/internal/transport/http/dto"

	"github.com/go-chi/chi/v5"
)

// MedalerHandler exposes medaler related HTTP handlers.
type MedalerHandler struct {
	service *service.Service
	logger  *slog.Logger
}

// NewMedalerHandler constructs a new handler.
func NewMedalerHandler(service *service.Service, logger *slog.Logger) *MedalerHandler {
	return &MedalerHandler{service: service, logger: logger}
}

// RegisterRoutes attaches the handler routes to the router group.
func (h *MedalerHandler) RegisterRoutes(r chi.Router) {
	r.Post("/medalers", h.create)
	r.Get("/medalers", h.list)
	r.Route("/medalers/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})
}

// @Summary Create a new medaler
// @Description Create a new medaler with the provided details
// @Tags Medaler
// @Accept json
// @Produce json
// @Param input body dto.CreateMedalerInput true "Medaler creation input"
// @Success 201 {object} dto.MedalerDTO
// @Failure 400 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /medalers [post]
func (h *MedalerHandler) create(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateMedalerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Create(r.Context(), input)
	if err != nil {
		if RespondValidationWithJSONTags(w, input, err) {
			return
		}
		h.logger.Error("create medaler", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to create medaler")
		return
	}

	respondJSON(w, http.StatusCreated, dto)
}

// @Summary Get all medalers
// @Description Retrieve a list of all medalers with pagination and filtering
// @Tags Medaler
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Param q query string false "Search query"
// @Param sort query string false "Sort order (e.g., id asc, name desc)"
// @Success 200 {object} dto.MedalerListResponse
// @Failure 500 {object} dto.Problem
// @Router /medalers [get]
func (h *MedalerHandler) list(w http.ResponseWriter, r *http.Request) {
	params := web.ParsePaginationParams(r)

	input := service.ListInput{
		Page:  params.Page,
		Limit: params.Limit,
		Query: params.Query,
		Sort:  params.Sort,
	}

	medalers, meta, err := h.service.List(r.Context(), input)
	if err != nil {
		h.logger.Error("list medalers", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list medalers")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"data": medalers,
		"meta": meta,
	})
}

// @Summary Get a medaler by ID
// @Description Retrieve a single medaler by its ID
// @Tags Medaler
// @Produce json
// @Param id path int true "Medaler ID"
// @Success 200 {object} dto.MedalerDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /medalers/{id} [get]
func (h *MedalerHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid medaler id")
		return
	}

	dto, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errors.New("medaler not found")) { // Using the generic error from repository for now
			respondError(w, http.StatusNotFound, "medaler not found")
			return
		}
		h.logger.Error("get medaler", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to get medaler")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}

// @Summary Update a medaler
// @Description Update an existing medaler with the provided details
// @Tags Medaler
// @Accept json
// @Produce json
// @Param id path int true "Medaler ID"
// @Param input body dto.UpdateMedalerInput true "Medaler update input"
// @Success 200 {object} dto.MedalerDTO
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /medalers/{id} [put]
func (h *MedalerHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid medaler id")
		return
	}

	var input dto.UpdateMedalerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	dto, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		if TryRespondValidation(w, err) {
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update medaler")
		return
	}

	respondJSON(w, http.StatusOK, dto)
}

// @Summary Delete a medaler
// @Description Delete a medaler by its ID
// @Tags Medaler
// @Produce json
// @Param id path int true "Medaler ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Problem
// @Failure 404 {object} dto.Problem
// @Failure 500 {object} dto.Problem
// @Router /medalers/{id} [delete]
func (h *MedalerHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := web.ParseUintParam(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid medaler id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		// In a real application, you would handle specific errors like ErrNotFound
		respondError(w, http.StatusInternalServerError, "failed to delete medaler")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
