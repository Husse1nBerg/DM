package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/dockworks/dm-web-backend/tests/testutil"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// RoleTestContext holds common test resources
type RoleTestContext struct {
	t        *testing.T
	handler  *handlers.RoleHandler
	recorder *httptest.ResponseRecorder
	ctx      echo.Context
	logger   *zap.SugaredLogger
}

// Helper function to convert a string to a pointer
func ptr(s string) *string {
	return &s
}

// setupRoleTest creates a new test context with Echo setup
func setupRoleTest(t *testing.T, method, url string, body []byte) *RoleTestContext {
	e := echo.New()
	e.Validator = &testutil.TestValidator{}

	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Initialize test logger
	log := logger.NewTestLogger()

	// Create a server with the test DB and logger
	s := &server.Server{
		DB:     testutil.InitTestDB(),
		Logger: log,
	}

	handler := handlers.NewRoleHandler(s)
	return &RoleTestContext{
		t:        t,
		handler:  handler,
		recorder: rec,
		ctx:      ctx,
		logger:   log.Zap,
	}
}

// createTestRoleRequest creates a request for a test role with unique values
func createTestRoleRequest(prefix string) requests.CreateRoleRequest {
	randomID := uuid.New().String()
	description := fmt.Sprintf("%s Role Description %s", prefix, randomID)
	isActive := true

	// Create a role with specific permissions using the new boolean model
	permissions := &models.Permissions{
		ReadUsers:          true,
		WriteUsers:         true,
		DeleteUsers:        false,
		ReadOrganizations:  true,
		WriteOrganizations: false,
		ReadMarinas:        true,
		WriteMarinas:       true,
		ReadRoles:          true,
		WriteRoles:         false,
		ReadSettings:       true,
		WriteSettings:      false,
	}

	return requests.CreateRoleRequest{
		Name:        fmt.Sprintf("%s Role %s", prefix, randomID),
		Description: &description,
		Permissions: permissions,
		IsActive:    &isActive,
	}
}

// createRoleViaAPI creates a role via API and returns its ID
func createRoleViaAPI(t *testing.T) uuid.UUID {
	// Create a test role request
	createReq := createTestRoleRequest("Create")
	createJSON, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create the role using the handler
	tc := setupRoleTest(t, http.MethodPost, "/role", createJSON)
	err = tc.handler.CreateRoleHandler(tc.ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, tc.recorder.Code)

	// Extract the role ID from the response
	var createResponse responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	roleData := createResponse.Data.(map[string]interface{})
	roleIDStr := roleData["id"].(string)
	roleID, err := uuid.Parse(roleIDStr)
	require.NoError(t, err)

	t.Logf("Created test role with ID: %s", roleID)
	return roleID
}

// TestCreateRole tests the role creation endpoint
func TestCreateRole(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		t.Run("Valid Role Creation", func(t *testing.T) {
			// Create a valid request with unique values
			req := createTestRoleRequest("Valid")
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupRoleTest(t, http.MethodPost, "/role", reqJSON)
			// tc.logger.Info("Testing valid role creation")

			// Call handler
			err = tc.handler.CreateRoleHandler(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusOK, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify data was created
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, req.Name, data["name"])
			if req.Description != nil {
				assert.Equal(t, *req.Description, data["description"])
			}

			// Verify permissions were saved
			permissions, ok := data["permissions"].(map[string]interface{})
			assert.True(t, ok)
			assert.NotNil(t, permissions)
		})

		t.Run("Invalid Request", func(t *testing.T) {
			// Create an invalid request (missing required name)
			req := createTestRoleRequest("Invalid")
			req.Name = "" // Name is required
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupRoleTest(t, http.MethodPost, "/role", reqJSON)
			// tc.logger.Info("Testing invalid role creation")

			// Call handler
			err = tc.handler.CreateRoleHandler(tc.ctx)
			assert.NoError(t, err)

			// Check response for validation error
			assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
		})
	})
}

// TestGetRoleEndpoints tests various get role endpoints
func TestGetRoleEndpoints(t *testing.T) {
	t.Run("Get By ID", func(t *testing.T) {
		// Create a test role first
		roleID := createRoleViaAPI(t)

		// Test getting by ID
		tc := setupRoleTest(t, http.MethodGet, "/role/:roleId", nil)
		tc.ctx.SetParamNames("roleId")
		tc.ctx.SetParamValues(roleID.String())
		// tc.logger.Info("Testing get role by ID", zap.String("id", roleID.String()))

		// Call handler
		err := tc.handler.GetRoleHandler(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify data exists
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, roleID.String(), data["id"])
	})

	t.Run("Get By ID Not Found", func(t *testing.T) {
		// Setup test with a non-existent ID
		nonExistentID := uuid.New()
		tc := setupRoleTest(t, http.MethodGet, "/role/:roleId", nil)
		tc.ctx.SetParamNames("roleId")
		tc.ctx.SetParamValues(nonExistentID.String())
		// tc.logger.Info("Testing get role by non-existent ID", zap.String("id", nonExistentID.String()))

		// Call handler
		err := tc.handler.GetRoleHandler(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusNotFound, tc.recorder.Code)
	})

	t.Run("Get By ID Invalid UUID", func(t *testing.T) {
		// Setup test with an invalid UUID
		tc := setupRoleTest(t, http.MethodGet, "/role/:roleId", nil)
		tc.ctx.SetParamNames("roleId")
		tc.ctx.SetParamValues("invalid-uuid")
		// tc.logger.Info("Testing get role with invalid UUID")

		// Call handler
		err := tc.handler.GetRoleHandler(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
	})

	t.Run("Get Role By Name", func(t *testing.T) {
		// Create a test role with a unique name
		req := createTestRoleRequest("NameTest")
		roleName := req.Name
		reqJSON, err := json.Marshal(req)
		require.NoError(t, err)

		// Create the role
		createTC := setupRoleTest(t, http.MethodPost, "/role", reqJSON)
		err = createTC.handler.CreateRoleHandler(createTC.ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, createTC.recorder.Code)

		// Test getting by name
		url := "/role/name"
		tc := setupRoleTest(t, http.MethodGet, url, nil)
		tc.ctx.QueryParams().Set("name", roleName)
		// tc.logger.Info("Testing get role by name", zap.String("name", roleName))

		// Call handler
		err = tc.handler.GetRoleByNameHandler(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify data exists and name matches
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, roleName, data["name"])
	})
}

// TestListRolesPaginated tests the paginated roles list endpoint
func TestListRolesPaginated(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// Create multiple test roles
		tc := setupRoleTest(t, http.MethodGet, "/role/list", nil)
		// tc.logger.Info("Creating multiple test roles for pagination test")

		// Create 5 roles
		for i := 0; i < 5; i++ {
			req := createTestRoleRequest(fmt.Sprintf("Page%d", i))
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			createTC := setupRoleTest(t, http.MethodPost, "/role", reqJSON)
			err = createTC.handler.CreateRoleHandler(createTC.ctx)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, createTC.recorder.Code)
		}

		// Test pagination with limit=3 and offset=0
		url := "/role/list?page=1&pageSize=3"
		tc = setupRoleTest(t, http.MethodGet, url, nil)
		tc.ctx.QueryParams().Set("page", "1")
		tc.ctx.QueryParams().Set("pageSize", "3")
		// tc.logger.Info("Testing paginated roles", zap.Int("page", 1), zap.Int("pageSize", 3))

		// Call handler
		err := tc.handler.ListRolesHandler(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify pagination data
		assert.GreaterOrEqual(t, response.Total, int64(5))
		assert.Equal(t, int32(3), response.PerPage)
		assert.Equal(t, int32(1), response.CurrentPage)

		// Verify response data is an array of roles
		data, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Len(t, data, 3)
	})
}

// TestUpdateRole tests the update role endpoint
func TestUpdateRole(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// Create a test role first
		roleID := createRoleViaAPI(t)

		// Test case 1: Full update
		t.Run("Full Update", func(t *testing.T) {
			// Create update request with all fields
			newDescription := "Updated Role Description - Full"
			newPermissions := &models.Permissions{
				ReadUsers:          true,
				WriteUsers:         false,
				DeleteUsers:        false,
				ReadOrganizations:  true,
				WriteOrganizations: false,
				ReadMarinas:        true,
				WriteMarinas:       false,
				ReadRoles:          true,
				WriteRoles:         false,
				ReadSettings:       true,
				WriteSettings:      false,
			}

			isActive := true
			uniqueName := fmt.Sprintf("Updated Role Name - Full %s", uuid.New().String())
			updateReq := requests.UpdateRoleRequest{
				Name:        ptr(uniqueName),
				Description: &newDescription,
				Permissions: newPermissions,
				IsActive:    &isActive,
			}

			updateJSON, err := json.Marshal(updateReq)
			require.NoError(t, err)

			// Setup update test
			tc := setupRoleTest(t, http.MethodPut, fmt.Sprintf("/role/%s", roleID), updateJSON)
			tc.ctx.SetParamNames("roleId")
			tc.ctx.SetParamValues(roleID.String())

			// Call update handler
			err = tc.handler.UpdateRoleHandler(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusOK, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify response data
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, uniqueName, data["name"])
			assert.Equal(t, *updateReq.Description, data["description"])

			// Verify permissions use the new boolean-based model
			permissions, ok := data["permissions"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, true, permissions["readUsers"])
			assert.Equal(t, false, permissions["writeUsers"])
			assert.Equal(t, true, permissions["readMarinas"])
			assert.Equal(t, false, permissions["writeMarinas"])
		})

		// Test case 2: Partial update (only description)
		t.Run("Partial Update - Description Only", func(t *testing.T) {
			// First get the current role to verify name is preserved
			currentRole, err := q.GetRoleByID(ctx, roleID)
			require.NoError(t, err)
			expectedName := currentRole.Name

			// Update only the description
			newPartialDescription := "Updated Role Description - Partial"
			partialUpdateReq := requests.UpdateRoleRequest{
				Description: &newPartialDescription,
			}

			partialUpdateJSON, err := json.Marshal(partialUpdateReq)
			require.NoError(t, err)

			// Setup update test
			tc := setupRoleTest(t, http.MethodPut, fmt.Sprintf("/role/%s", roleID), partialUpdateJSON)
			tc.ctx.SetParamNames("roleId")
			tc.ctx.SetParamValues(roleID.String())

			// Call update handler
			err = tc.handler.UpdateRoleHandler(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusOK, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify only description changed, name preserved
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, expectedName, data["name"])
			assert.Equal(t, newPartialDescription, data["description"])
		})

		// Test case 3: Permissions-only update
		t.Run("Permissions Only Update", func(t *testing.T) {
			// Update only the permissions
			adminPermissions := &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          true,
				DeleteUsers:         true,
				ReadOrganizations:   true,
				WriteOrganizations:  true,
				DeleteOrganizations: true,
				ReadMarinas:         true,
				WriteMarinas:        true,
				DeleteMarinas:       true,
				ReadRoles:           true,
				WriteRoles:          true,
				DeleteRoles:         true,
				ReadSettings:        true,
				WriteSettings:       true,
			}

			permissionsUpdateReq := requests.UpdateRoleRequest{
				Permissions: adminPermissions,
			}

			permUpdateJSON, err := json.Marshal(permissionsUpdateReq)
			require.NoError(t, err)

			// Setup update test
			tc := setupRoleTest(t, http.MethodPut, fmt.Sprintf("/role/%s", roleID), permUpdateJSON)
			tc.ctx.SetParamNames("roleId")
			tc.ctx.SetParamValues(roleID.String())

			// Call update handler
			err = tc.handler.UpdateRoleHandler(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusOK, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify permissions were updated to admin level
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)

			permissions, ok := data["permissions"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, true, permissions["writeUsers"])
			assert.Equal(t, true, permissions["deleteUsers"])
			assert.Equal(t, true, permissions["writeRoles"])
			assert.Equal(t, true, permissions["deleteRoles"])
		})

		// Test case 4: Invalid Role ID
		t.Run("Invalid Role ID", func(t *testing.T) {
			updateReq := requests.UpdateRoleRequest{
				Name: ptr("Test Role"),
			}

			updateJSON, err := json.Marshal(updateReq)
			require.NoError(t, err)

			// Setup update test with invalid UUID
			tc := setupRoleTest(t, http.MethodPut, "/role/invalid-uuid", updateJSON)
			tc.ctx.SetParamNames("roleId")
			tc.ctx.SetParamValues("invalid-uuid")

			// Call update handler
			err = tc.handler.UpdateRoleHandler(tc.ctx)
			assert.NoError(t, err)

			// Check response is a 400 error
			assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "Invalid role ID format", response.Error)
		})
	})
}

// TestDeleteRole tests the delete role endpoint
func TestDeleteRole(t *testing.T) {
	// Create a test role first
	roleID := createRoleViaAPI(t)

	// Setup delete test
	tc := setupRoleTest(t, http.MethodDelete, "/role/:roleId", nil)
	tc.ctx.SetParamNames("roleId")
	tc.ctx.SetParamValues(roleID.String())
	// tc.logger.Info("Testing delete role", zap.String("id", roleID.String()))

	// Call delete handler
	err := tc.handler.DeleteRoleHandler(tc.ctx)
	assert.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, tc.recorder.Code)

	var response responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify success message
	assert.Equal(t, "Role successfully deleted", response.Message)

	// Attempt to get the deleted role (should be soft-deleted)
	getTC := setupRoleTest(t, http.MethodGet, "/role/:roleId", nil)
	getTC.ctx.SetParamNames("roleId")
	getTC.ctx.SetParamValues(roleID.String())
	// getTC.logger.Info("Testing get on deleted role", zap.String("id", roleID.String()))

	// This should fail since the role is soft-deleted
	err = getTC.handler.GetRoleHandler(getTC.ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getTC.recorder.Code)
}
