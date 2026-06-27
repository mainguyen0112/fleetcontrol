package satellite

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	sat := &Satellite{
		Name:   req.Name,
		Region: req.Region,
	}

	created, err := h.service.Create(r.Context(), sat)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	sats, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sats)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid satellite id")
		return
	}

	sat, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_FAILED", err.Error())
		return
	}
	if sat == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "satellite not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sat)
}

type updateRequest struct {
	Region string `json:"region"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid satellite id")
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	fromOperator := r.Header.Get("X-FleetControl-Operator") == "true"

	updated, err := h.service.Update(r.Context(), id, req.Region, fromOperator)
	if err != nil {
		if errors.Is(err, ErrManagedByOperator) {
			writeError(w, http.StatusForbidden, "SATELLITE_MANAGED_BY_OPERATOR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "satellite not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid satellite id")
		return
	}

	fromOperator := r.Header.Get("X-FleetControl-Operator") == "true"

	if err := h.service.Delete(r.Context(), id, fromOperator); err != nil {
		if errors.Is(err, ErrManagedByOperator) {
			writeError(w, http.StatusForbidden, "SATELLITE_MANAGED_BY_OPERATOR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid satellite id")
		return
	}

	updated, err := h.service.Heartbeat(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "HEARTBEAT_FAILED", err.Error())
		return
	}
	if updated == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "satellite not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}
