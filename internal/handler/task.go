package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"taskmanager/internal/repository"
	"taskmanager/internal/task"
)

const maxRequestBodyBytes = 1 << 20

type TaskHandler struct {
	repository repository.TaskRepository
}

func NewTaskHandler(repository repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repository: repository}
}

func (h *TaskHandler) RegisterRoutes(router chi.Router) {
	router.Post("/tasks", h.create)
	router.Get("/tasks", h.list)
	router.Get("/tasks/{id}", h.get)
	router.Put("/tasks/{id}", h.update)
	router.Patch("/tasks/{id}", h.patch)
	router.Delete("/tasks/{id}", h.delete)
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if strings.TrimSpace(request.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	item, err := h.repository.Create(r.Context(), task.CreateInput{
		Title:       strings.TrimSpace(request.Title),
		Description: request.Description,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create task")
		return
	}
	w.Header().Set("Location", "/tasks/"+strconv.FormatInt(item.ID, 10))
	writeJSON(w, http.StatusCreated, item)
}

func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repository.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list tasks")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *TaskHandler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}
	item, err := h.repository.Get(r.Context(), id)
	if err != nil {
		h.writeRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}
	var request struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Completed   *bool   `json:"completed"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Title == nil || request.Description == nil || request.Completed == nil {
		writeError(w, http.StatusBadRequest, "title, description, and completed are required for a full update")
		return
	}
	title := strings.TrimSpace(*request.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	item, err := h.repository.Update(r.Context(), id, task.UpdateInput{
		Title:       title,
		Description: *request.Description,
		Completed:   *request.Completed,
	})
	if err != nil {
		h.writeRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TaskHandler) patch(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}
	var request struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Completed   *bool   `json:"completed"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.Title == nil && request.Description == nil && request.Completed == nil {
		writeError(w, http.StatusBadRequest, "at least one task field must be provided")
		return
	}
	if request.Title != nil {
		title := strings.TrimSpace(*request.Title)
		if title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		request.Title = &title
	}

	item, err := h.repository.Patch(r.Context(), id, task.PatchInput{
		Title:       request.Title,
		Description: request.Description,
		Completed:   request.Completed,
	})
	if err != nil {
		h.writeRepositoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TaskHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}
	if err := h.repository.Delete(r.Context(), id); err != nil {
		h.writeRepositoryError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseTaskID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "task id must be a positive integer")
		return 0, false
	}
	return id, true
}

func (h *TaskHandler) writeRepositoryError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "could not process task")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "request body must contain valid JSON")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
