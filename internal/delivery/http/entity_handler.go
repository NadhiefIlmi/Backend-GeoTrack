package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"geotrack/backend/internal/domain"
	"geotrack/backend/pkg/validator"
)

type EntityHandler struct {
	usecase domain.EntityUsecase
}

// NewEntityHandler creates a new HTTP delivery handler for entities
func NewEntityHandler(usecase domain.EntityUsecase) *EntityHandler {
	return &EntityHandler{usecase: usecase}
}

func (h *EntityHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := domain.EntityFilter{
		Search: strings.TrimSpace(q.Get("search")),
		SortBy: strings.TrimSpace(q.Get("sort_by")),
		Order:  strings.TrimSpace(q.Get("order")),
		Page:   page,
		Limit:  limit,
	}

	rawType := strings.TrimSpace(q.Get("type"))
	if rawType != "" {
		t := domain.EntityType(rawType)
		filter.Type = &t
	}

	rawStatus := strings.TrimSpace(q.Get("status"))
	if rawStatus != "" {
		s := domain.EntityStatus(rawStatus)
		filter.Status = &s
	}

	entities, meta, err := h.usecase.List(r.Context(), filter)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve entities", nil)
		return
	}

	paginationMeta := &MetaPagination{
		Page:         meta.Page,
		Limit:        meta.Limit,
		TotalRecords: meta.TotalRecords,
		TotalPages:   meta.TotalPages,
	}

	SendSuccess(w, http.StatusOK, "Entities retrieved successfully", entities, paginationMeta)
}

func (h *EntityHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		SendError(w, http.StatusBadRequest, "INVALID_ID", "Entity ID is required", nil)
		return
	}

	entity, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			SendError(w, http.StatusNotFound, "ENTITY_NOT_FOUND", "Entity was not found", nil)
			return
		}
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve entity", nil)
		return
	}

	SendSuccess(w, http.StatusOK, "Entity retrieved successfully", entity, nil)
}

func (h *EntityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body", nil)
		return
	}

	// Structural & semantic validation
	if fieldErrors := validator.ValidateStruct(req); len(fieldErrors) > 0 {
		SendError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Input validation failed", fieldErrors)
		return
	}

	entity, err := h.usecase.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateCode) {
			SendError(w, http.StatusConflict, "DUPLICATE_CODE", "Entity code is already registered", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidCoordinates) {
			SendError(w, http.StatusUnprocessableEntity, "INVALID_COORDINATES", "Latitude must be between [-90, 90] and Longitude between [-180, 180]", nil)
			return
		}
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create entity", nil)
		return
	}

	SendSuccess(w, http.StatusCreated, "Entity created successfully", entity, nil)
}

func (h *EntityHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		SendError(w, http.StatusBadRequest, "INVALID_ID", "Entity ID is required", nil)
		return
	}

	var req domain.UpdateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "INVALID_JSON", "Malformed JSON request body", nil)
		return
	}

	if fieldErrors := validator.ValidateStruct(req); len(fieldErrors) > 0 {
		SendError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Input validation failed", fieldErrors)
		return
	}

	entity, err := h.usecase.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			SendError(w, http.StatusNotFound, "ENTITY_NOT_FOUND", "Entity was not found", nil)
			return
		}
		if errors.Is(err, domain.ErrDuplicateCode) {
			SendError(w, http.StatusConflict, "DUPLICATE_CODE", "Entity code is already in use by another entity", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidCoordinates) {
			SendError(w, http.StatusUnprocessableEntity, "INVALID_COORDINATES", "Latitude must be between [-90, 90] and Longitude between [-180, 180]", nil)
			return
		}
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update entity", nil)
		return
	}

	SendSuccess(w, http.StatusOK, "Entity updated successfully", entity, nil)
}

func (h *EntityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		SendError(w, http.StatusBadRequest, "INVALID_ID", "Entity ID is required", nil)
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			SendError(w, http.StatusNotFound, "ENTITY_NOT_FOUND", "Entity was not found", nil)
			return
		}
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete entity", nil)
		return
	}

	SendSuccess(w, http.StatusOK, "Entity deleted successfully", nil, nil)
}

