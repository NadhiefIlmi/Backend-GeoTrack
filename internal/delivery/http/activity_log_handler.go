package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"geotrack/backend/internal/domain"
)

type ActivityLogHandler struct {
	usecase domain.ActivityLogUsecase
}

// NewActivityLogHandler creates a new HTTP delivery handler for activity logs
func NewActivityLogHandler(usecase domain.ActivityLogUsecase) *ActivityLogHandler {
	return &ActivityLogHandler{usecase: usecase}
}

func (h *ActivityLogHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := domain.ActivityFilter{
		Search:   strings.TrimSpace(q.Get("search")),
		EntityID: strings.TrimSpace(q.Get("entity_id")),
		Page:     page,
		Limit:    limit,
	}

	rawAction := strings.TrimSpace(q.Get("action"))
	if rawAction != "" {
		a := domain.ActivityAction(rawAction)
		filter.Action = &a
	}

	logs, meta, err := h.usecase.List(r.Context(), filter)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve activity logs", nil)
		return
	}

	paginationMeta := &MetaPagination{
		Page:         meta.Page,
		Limit:        meta.Limit,
		TotalRecords: meta.TotalRecords,
		TotalPages:   meta.TotalPages,
	}

	SendSuccess(w, http.StatusOK, "Activity logs retrieved successfully", logs, paginationMeta)
}

func (h *ActivityLogHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		SendError(w, http.StatusBadRequest, "INVALID_ID", "Activity log ID is required", nil)
		return
	}

	logEntry, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrActivityNotFound) {
			SendError(w, http.StatusNotFound, "ACTIVITY_NOT_FOUND", "Activity log was not found", nil)
			return
		}
		SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve activity log", nil)
		return
	}

	SendSuccess(w, http.StatusOK, "Activity log retrieved successfully", logEntry, nil)
}

