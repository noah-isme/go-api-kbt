package handler

import (
    "encoding/json"
    "log/slog"
    "net/http"
    "strconv"

    service "go-api-kbt/internal/service/event"

    "github.com/go-chi/chi/v5"
)

type EventHandler struct {
    service *service.Service
    logger  *slog.Logger
}

func NewEventHandler(s *service.Service, logger *slog.Logger) *EventHandler { return &EventHandler{service: s, logger: logger} }

func (h *EventHandler) RegisterRoutes(r chi.Router) {
    r.Post("/events", h.create)
    r.Get("/events", h.list)
    r.Route("/events/{id}", func(r chi.Router) {
        r.Get("/", h.get)
        r.Put("/", h.update)
        r.Delete("/", h.delete)
    })
}

func (h *EventHandler) create(w http.ResponseWriter, r *http.Request) {
    var in service.CreateInput
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request payload")
        return
    }
    dto, err := h.service.Create(r.Context(), in)
    if err != nil {
        if respondValidationWithJSONTags(w, in, err) { return }
        h.logger.Error("create event", slog.String("error", err.Error()))
        respondError(w, http.StatusInternalServerError, "failed to create event")
        return
    }
    respondJSON(w, http.StatusCreated, dto)
}

func (h *EventHandler) list(w http.ResponseWriter, r *http.Request) {
    limit, offset := parsePagination(r)
    dtos, err := h.service.List(r.Context(), limit, offset)
    if err != nil {
        h.logger.Error("list events", slog.String("error", err.Error()))
        respondError(w, http.StatusInternalServerError, "failed to list events")
        return
    }
    respondJSON(w, http.StatusOK, map[string]any{"data": dtos, "limit": limit, "offset": offset})
}

func (h *EventHandler) get(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
    dto, err := h.service.Get(r.Context(), uint(id))
    if err != nil {
        respondError(w, http.StatusNotFound, "event not found")
        return
    }
    respondJSON(w, http.StatusOK, dto)
}

func (h *EventHandler) update(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
    var in service.UpdateInput
    if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request payload")
        return
    }
    dto, err := h.service.Update(r.Context(), uint(id), in)
    if err != nil {
        if respondValidationWithJSONTags(w, in, err) { return }
        respondError(w, http.StatusInternalServerError, "failed to update event")
        return
    }
    respondJSON(w, http.StatusOK, dto)
}

func (h *EventHandler) delete(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
    if err := h.service.Delete(r.Context(), uint(id)); err != nil {
        respondError(w, http.StatusInternalServerError, "failed to delete event")
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
