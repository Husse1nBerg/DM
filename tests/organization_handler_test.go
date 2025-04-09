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
	"github.com/dockworks/dm-web-backend/tests/testutil"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestContext holds common test resources
type TestContext struct {
	t        *testing.T
	handler  *handlers.OrganizationHandler
	recorder *httptest.ResponseRecorder
	ctx      echo.Context
	logger   *zap.SugaredLogger
}

// setupTest creates a new test context with Echo setup
func setupTest(t *testing.T, method, url string, body []byte) *TestContext {
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

	handler := handlers.NewOrganizationHandler(s)
	return &TestContext{
		t:        t,
		handler:  handler,
		recorder: rec,
		ctx:      ctx,
		logger:   log.Zap,
	}
}

// createTestOrgRequest creates a request for a test organization with unique values
func createTestOrgRequest(prefix string) requests.CreateOrganizationRequest {
	randomID := uuid.New().String()
	return requests.CreateOrganizationRequest{
		Name:     fmt.Sprintf("%s Org %s", prefix, randomID),
		Email:    fmt.Sprintf("%s-%s@example.com", prefix, randomID),
		Website:  &[]string{"https://test.com"}[0],
		Country:  &[]string{"United States"}[0],
		Phone:    &[]string{"+1234567890"}[0],
		IsActive: true,
		IsTest:   true,
		Address: &requests.CreateAddressRequest{
			Street:     &[]string{"123 Test St"}[0],
			City:       &[]string{"Test City"}[0],
			State:      &[]string{"Test State"}[0],
			PostalCode: &[]string{"12345"}[0],
			Country:    &[]string{"USA"}[0],
			Latitude:   &[]float64{40.7128}[0],
			Longitude:  &[]float64{-74.0060}[0],
		},
	}
}

// createOrganizationViaAPI creates an organization via API and returns its ID
func createOrganizationViaAPI(t *testing.T) uuid.UUID {
	// Create a test organization request
	createReq := createTestOrgRequest("Create")
	createJSON, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create the organization using the handler
	tc := setupTest(t, http.MethodPost, "/organizations", createJSON)
	err = tc.handler.CreateOrganization(tc.ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, tc.recorder.Code)

	// Extract the organization ID from the response
	var createResponse responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	orgData := createResponse.Data.(map[string]interface{})
	orgIDStr := orgData["id"].(string)
	orgID, err := uuid.Parse(orgIDStr)
	require.NoError(t, err)

	t.Logf("Created test organization with ID: %s", orgID)
	return orgID
}

// TestCreateOrganization tests the organization creation endpoint
func TestCreateOrganization(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		t.Run("Valid Organization Creation", func(t *testing.T) {
			// Create a valid request with unique values
			req := createTestOrgRequest("Valid")
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupTest(t, http.MethodPost, "/organizations", reqJSON)
			// tc.logger.Info("Testing valid organization creation")

			// Call handler
			err = tc.handler.CreateOrganization(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusCreated, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify data was created
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, req.Name, data["name"])
			assert.Equal(t, req.Email, data["email"])
		})

		t.Run("Invalid Request", func(t *testing.T) {
			// Create an invalid request (missing required name)
			req := createTestOrgRequest("Invalid")
			req.Name = "" // Name is required
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupTest(t, http.MethodPost, "/organizations", reqJSON)
			// tc.logger.Info("Testing invalid organization creation")

			// Call handler
			err = tc.handler.CreateOrganization(tc.ctx)
			assert.NoError(t, err)

			// Check response for validation error
			assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
		})
	})
}

// TestGetOrganizationEndpoints tests various get organization endpoints
func TestGetOrganizationEndpoints(t *testing.T) {
	t.Run("Get By ID", func(t *testing.T) {
		// Create a test organization first
		orgID := createOrganizationViaAPI(t)

		// Test getting by ID
		tc := setupTest(t, http.MethodGet, "/organizations/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(orgID.String())
		// tc.logger.Info("Testing get organization by ID", zap.String("id", orgID.String()))

		// Call handler
		err := tc.handler.GetOrganizationByID(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify data exists
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, orgID.String(), data["id"])
	})

	t.Run("Get By ID Not Found", func(t *testing.T) {
		// Setup test with a non-existent ID
		nonExistentID := uuid.New()
		tc := setupTest(t, http.MethodGet, "/organizations/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(nonExistentID.String())
		// tc.logger.Info("Testing get organization by non-existent ID", zap.String("id", nonExistentID.String()))

		// Call handler
		err := tc.handler.GetOrganizationByID(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusNotFound, tc.recorder.Code)
	})

	t.Run("Get By ID Invalid UUID", func(t *testing.T) {
		// Setup test with an invalid UUID
		tc := setupTest(t, http.MethodGet, "/organizations/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues("invalid-uuid")
		// tc.logger.Info("Testing get organization with invalid UUID")

		// Call handler
		err := tc.handler.GetOrganizationByID(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
	})

	t.Run("Get With Address", func(t *testing.T) {
		// Create a test organization first
		orgID := createOrganizationViaAPI(t)

		// Test getting with address
		tc := setupTest(t, http.MethodGet, "/organizations/:id/with-address", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(orgID.String())
		// tc.logger.Info("Testing get organization with address", zap.String("id", orgID.String()))

		// Call handler
		err := tc.handler.GetOrganizationWithAddress(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify organization and address data exists
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, orgID.String(), data["id"])

		// Check for address
		address, ok := data["address"].(map[string]interface{})
		assert.True(t, ok, "Address should be included in response")
		assert.NotEmpty(t, address["city"])
	})
}

// TestGetOrganizationByEmail tests the get organization by email endpoint
func TestGetOrganizationByEmail(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// Create a test organization with a unique email
		randomID := uuid.New().String()
		email := fmt.Sprintf("email-test-%s@example.com", randomID)
		req := createTestOrgRequest("Email")
		req.Email = email

		reqJSON, err := json.Marshal(req)
		require.NoError(t, err)

		// Create the organization
		createTC := setupTest(t, http.MethodPost, "/organizations", reqJSON)
		createTC.logger.Info("Creating test organization for email test", zap.String("email", email))
		err = createTC.handler.CreateOrganization(createTC.ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, createTC.recorder.Code)

		// Now test getting the organization by email
		url := fmt.Sprintf("/organizations/by-email?email=%s", email)
		tc := setupTest(t, http.MethodGet, url, nil)
		tc.ctx.QueryParams().Set("email", email)
		// tc.logger.Info("Testing get organization by email", zap.String("email", email))

		// Call handler
		err = tc.handler.GetOrganizationByEmail(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify data
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, email, data["email"])
	})
}

// TestGetOrganizationsPaginated tests the paginated organizations endpoint
func TestGetOrganizationsPaginated(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// Create multiple test organizations
		tc := setupTest(t, http.MethodGet, "/organizations", nil)
		// tc.logger.Info("Creating multiple test organizations for pagination test")

		// Create 5 organizations
		for i := 0; i < 5; i++ {
			req := createTestOrgRequest(fmt.Sprintf("Page%d", i))
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			createTC := setupTest(t, http.MethodPost, "/organizations", reqJSON)
			err = createTC.handler.CreateOrganization(createTC.ctx)
			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, createTC.recorder.Code)
		}

		// Test pagination with limit=3 and offset=0
		url := "/organizations?page=1&pageSize=3"
		tc = setupTest(t, http.MethodGet, url, nil)
		tc.ctx.QueryParams().Set("page", "1")
		tc.ctx.QueryParams().Set("pageSize", "3")
		// tc.logger.Info("Testing paginated organizations", zap.Int("page", 1), zap.Int("pageSize", 3))

		// Call handler
		err := tc.handler.GetOrganizationsPaginated(tc.ctx)
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

		// Verify response data is an array of organizations
		data, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Len(t, data, 3)
	})
}

// TestUpdateOrganizationEndpoints tests the update organization endpoints
func TestUpdateOrganizationEndpoints(t *testing.T) {
	t.Run("Update Organization", func(t *testing.T) {
		// Create a test organization first
		orgID := createOrganizationViaAPI(t)

		// Create update request
		newCountry := "New Country"
		updateReq := requests.UpdateOrganizationRequest{
			Country: &newCountry,
		}

		updateJSON, err := json.Marshal(updateReq)
		require.NoError(t, err)

		// Setup update test
		tc := setupTest(t, http.MethodPut, "/organizations/:id", updateJSON)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(orgID.String())
		// tc.logger.Info("Testing update organization", zap.String("id", orgID.String()))

		// Call update handler
		err = tc.handler.UpdateOrganization(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response data
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, newCountry, data["country"])
	})

	t.Run("Update Organization Address", func(t *testing.T) {
		// Create a test organization first
		orgID := createOrganizationViaAPI(t)

		// Create update request with address changes
		newCity := "New City"
		updateReq := requests.UpdateOrgAddressRequest{
			Address: &requests.UpdateAddressRequest{
				City: &newCity,
			},
		}

		updateJSON, err := json.Marshal(updateReq)
		require.NoError(t, err)

		// Setup update test
		tc := setupTest(t, http.MethodPut, "/organizations/:id/with-address", updateJSON)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(orgID.String())
		// tc.logger.Info("Testing update organization address", zap.String("id", orgID.String()))

		// Call update handler
		err = tc.handler.UpdateOrgAddress(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response data
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		// Verify address was updated
		address, ok := data["address"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, newCity, address["city"])
	})
}

// TestDeleteOrganization tests the delete organization endpoint
func TestDeleteOrganization(t *testing.T) {
	// Create a test organization first
	orgID := createOrganizationViaAPI(t)

	// Setup delete test
	tc := setupTest(t, http.MethodDelete, "/organizations/:id", nil)
	tc.ctx.SetParamNames("id")
	tc.ctx.SetParamValues(orgID.String())
	// tc.logger.Info("Testing delete organization", zap.String("id", orgID.String()))

	// Call delete handler
	err := tc.handler.DeleteOrganization(tc.ctx)
	assert.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, tc.recorder.Code)

	var response responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify success message
	assert.Equal(t, "Organization successfully deleted", response.Message)

	// Attempt to get the deleted organization (should be soft-deleted)
	getTC := setupTest(t, http.MethodGet, "/organizations/:id", nil)
	getTC.ctx.SetParamNames("id")
	getTC.ctx.SetParamValues(orgID.String())
	// getTC.logger.Info("Testing get on deleted organization", zap.String("id", orgID.String()))

	// This should fail since the organization is soft-deleted
	err = getTC.handler.GetOrganizationByID(getTC.ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getTC.recorder.Code)
}
