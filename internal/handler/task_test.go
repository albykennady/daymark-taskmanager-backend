package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"taskmanager/internal/database"
	"taskmanager/internal/repository"
	"taskmanager/internal/task"
)

func TestTaskCRUD(t *testing.T) {
	db, err := database.OpenSQLiteMemory(context.Background())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	if err := database.SeedTasks(context.Background(), db); err != nil {
		t.Fatalf("seed test database: %v", err)
	}

	router := chi.NewRouter()
	NewTaskHandler(repository.NewSQLiteRepository(db)).RegisterRoutes(router)

	putResponse := performRequest(t, router, http.MethodPut, "/tasks/1", `{"title":"Replace seeded task","description":"Updated through PUT","completed":true}`)
	if putResponse.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want %d: %s", putResponse.Code, http.StatusOK, putResponse.Body)
	}
	var putTask task.Task
	if err := json.Unmarshal(putResponse.Body.Bytes(), &putTask); err != nil {
		t.Fatalf("decode PUT response: %v", err)
	}
	if putTask.ID != 1 || putTask.Title != "Replace seeded task" || !putTask.Completed {
		t.Fatalf("PUT did not update task 1: %+v", putTask)
	}

	createResponse := performRequest(t, router, http.MethodPost, "/tasks", `{"title":"Plan sprint","description":"Prepare tickets"}`)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", createResponse.Code, http.StatusCreated, createResponse.Body)
	}
	var created task.Task
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID < 1 || created.Title != "Plan sprint" {
		t.Fatalf("unexpected created task: %+v", created)
	}

	listResponse := performRequest(t, router, http.MethodGet, "/tasks", "")
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", listResponse.Code, http.StatusOK, listResponse.Body)
	}
	var listed []task.Task
	if err := json.Unmarshal(listResponse.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listed) != 4 {
		t.Fatalf("list did not return created task: %+v", listed)
	}
	foundCreated := false
	for _, item := range listed {
		if item.ID == created.ID {
			foundCreated = true
			break
		}
	}
	if !foundCreated {
		t.Fatalf("list omitted created task %d: %+v", created.ID, listed)
	}

	createdID := strconv.FormatInt(created.ID, 10)
	updateResponse := performRequest(t, router, http.MethodPut, "/tasks/"+createdID, `{"title":"Plan next sprint","description":"Prioritize tickets","completed":false}`)
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", updateResponse.Code, http.StatusOK, updateResponse.Body)
	}

	patchResponse := performRequest(t, router, http.MethodPatch, "/tasks/"+createdID, `{"completed":true}`)
	if patchResponse.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want %d: %s", patchResponse.Code, http.StatusOK, patchResponse.Body)
	}

	getResponse := performRequest(t, router, http.MethodGet, "/tasks/"+createdID, "")
	var updated task.Task
	if err := json.Unmarshal(getResponse.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if !updated.Completed || updated.Title != "Plan next sprint" || updated.Description != "Prioritize tickets" {
		t.Fatalf("patch did not preserve task fields: %+v", updated)
	}

	deleteResponse := performRequest(t, router, http.MethodDelete, "/tasks/"+createdID, "")
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteResponse.Code, http.StatusNoContent)
	}
	getDeletedResponse := performRequest(t, router, http.MethodGet, "/tasks/"+createdID, "")
	if getDeletedResponse.Code != http.StatusNotFound {
		t.Fatalf("get deleted status = %d, want %d", getDeletedResponse.Code, http.StatusNotFound)
	}
}

func TestTaskEndpointRejectsNonIntegerID(t *testing.T) {
	db, err := database.OpenSQLiteMemory(context.Background())
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	router := chi.NewRouter()
	NewTaskHandler(repository.NewSQLiteRepository(db)).RegisterRoutes(router)

	response := performRequest(t, router, http.MethodPut, "/tasks/not-an-id", `{"title":"Updated","description":"Body","completed":true}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body)
	}
}

func TestCreateRequiresTitle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := database.OpenSQLiteMemory(ctx)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	router := chi.NewRouter()
	NewTaskHandler(repository.NewSQLiteRepository(db)).RegisterRoutes(router)

	response := performRequest(t, router, http.MethodPost, "/tasks", `{"description":"Missing a title"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func performRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
