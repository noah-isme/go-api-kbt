package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	service "go-api-kbt/internal/service/location"

	"github.com/go-chi/chi/v5"
)

type LocationHandler struct {
	service *service.Service
	logger  *slog.Logger
}

func NewLocationHandler(s *service.Service, logger *slog.Logger) *LocationHandler {
	return &LocationHandler{service: s, logger: logger}
}

func (h *LocationHandler) RegisterRoutes(r chi.Router) {
	r.Post("/locations", h.create)
	r.Get("/locations", h.list)
	r.Route("/locations/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
	})
}

func (h *LocationHandler) create(w http.ResponseWriter, r *http.Request) {
	var in service.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	dto, err := h.service.Create(r.Context(), in)
	if err != nil {
		if respondValidationWithJSONTags(w, in, err) {
			return
		}
		h.logger.Error("create location", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to create location")
		return
	}
	respondJSON(w, http.StatusCreated, dto)
}

func (h *LocationHandler) list(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	dtos, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("list locations", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to list locations")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": dtos, "limit": limit, "offset": offset})
}

func (h *LocationHandler) get(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	dto, err := h.service.Get(r.Context(), uint(id))
	if err != nil {
		respondError(w, http.StatusNotFound, "location not found")
		return
	}
	respondJSON(w, http.StatusOK, dto)
}

func (h *LocationHandler) update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	var in service.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	dto, err := h.service.Update(r.Context(), uint(id), in)
	if err != nil {
		if respondValidationWithJSONTags(w, in, err) {
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update location")
		return
	}
	respondJSON(w, http.StatusOK, dto)
}

func (h *LocationHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err := h.service.Delete(r.Context(), uint(id)); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete location")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
