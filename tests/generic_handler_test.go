package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not initialized")
	}

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a server with the test DB
	s := &server.Server{
		DB: testDB,
	}

	h := handlers.NewGenericHandler(s)

	// Execute the handler
	if err := h.HealthHandler(c); err != nil {
		t.Fatalf("HealthHandler returned error: %v", err)
	}

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response responses.BaseResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check response data
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Response data should be a map")
	assert.Equal(t, "up", data["status"], "Database status should be 'up'")
}

func TestProjectDetailsHandler(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/project-details", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a server with the test DB (even though this handler doesn't use it)
	s := &server.Server{
		DB: testDB,
	}

	h := handlers.NewGenericHandler(s)

	// Execute the handler
	if err := h.ProjectDetailsHandler(c); err != nil {
		t.Fatalf("ProjectDetailsHandler returned error: %v", err)
	}

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response responses.BaseResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check response data
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Response data should be a map")
	assert.Equal(t, "Marina Management System", data["name"])
	assert.Equal(t, "1.0.0", data["version"])
	assert.Equal(t, "A multi-tenant platform for marina management", data["description"])
	assert.Equal(t, "Go, Echo, PostgreSQL, SQLC", data["tech_stack"])
}
