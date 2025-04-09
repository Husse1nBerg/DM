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

// MarinaTestContext holds common test resources
type MarinaTestContext struct {
	t        *testing.T
	handler  *handlers.MarinaHandler
	recorder *httptest.ResponseRecorder
	ctx      echo.Context
	logger   *zap.SugaredLogger
}

// setupMarinaTest creates a new test context with Echo setup
func setupMarinaTest(t *testing.T, method, url string, body []byte) *MarinaTestContext {
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

	handler := handlers.NewMarinaHandler(s)
	return &MarinaTestContext{
		t:        t,
		handler:  handler,
		recorder: rec,
		ctx:      ctx,
		logger:   log.Zap,
	}
}

// createTestMarinaRequest creates a request for a test marina with unique values
func createTestMarinaRequest(prefix string, orgID uuid.UUID) requests.CreateMarinaRequest {
	randomID := uuid.New().String()
	maxUsers := int32(100)
	isActive := true
	isTest := true

	return requests.CreateMarinaRequest{
		OrganizationID: orgID,
		Name:           fmt.Sprintf("%s Marina %s", prefix, randomID),
		Email:          fmt.Sprintf("%s-marina-%s@example.com", prefix, randomID),
		Location:       &[]string{"Miami Beach"}[0],
		Phone:          &[]string{"+1987654321"}[0],
		Country:        &[]string{"United States"}[0],
		Currency:       &[]string{"USD"}[0],
		WorkingHours:   models.DefaultWorkingHours(),
		Website:        &[]string{"https://marina-test.com"}[0],
		MaxUsers:       &maxUsers,
		IsActive:       &isActive,
		IsTest:         &isTest,
		Address: &requests.CreateAddressRequest{
			Street:     &[]string{"456 Marina Ave"}[0],
			City:       &[]string{"Marina City"}[0],
			State:      &[]string{"Florida"}[0],
			PostalCode: &[]string{"33139"}[0],
			Country:    &[]string{"USA"}[0],
			Latitude:   &[]float64{25.7617}[0],
			Longitude:  &[]float64{-80.1918}[0],
		},
	}
}

// createOrganizationForTest creates an organization for test purposes and returns its ID
func createOrganizationForTest(t *testing.T) uuid.UUID {
	// Create a test organization request
	createReq := createTestOrgRequest("MarinaTest")
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

// createMarinaViaAPI creates a marina via API and returns its ID
func createMarinaViaAPI(t *testing.T, orgID uuid.UUID) uuid.UUID {
	// Create a test marina request
	createReq := createTestMarinaRequest("Create", orgID)
	createJSON, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create the marina using the handler
	tc := setupMarinaTest(t, http.MethodPost, "/marinas", createJSON)
	err = tc.handler.CreateMarina(tc.ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, tc.recorder.Code)

	// Extract the marina ID from the response
	var createResponse responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	marinaData := createResponse.Data.(map[string]interface{})
	marinaIDStr := marinaData["id"].(string)
	marinaID, err := uuid.Parse(marinaIDStr)
	require.NoError(t, err)

	t.Logf("Created test marina with ID: %s", marinaID)
	return marinaID
}

// TestCreateMarina tests the marina creation endpoint
func TestCreateMarina(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// First create an organization for the marina
		orgID := createOrganizationForTest(t)

		t.Run("Valid Marina Creation", func(t *testing.T) {
			// Create a valid request with unique values
			req := createTestMarinaRequest("Valid", orgID)
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupMarinaTest(t, http.MethodPost, "/marinas", reqJSON)
			// tc.logger.Info("Testing valid marina creation")

			// Call handler
			err = tc.handler.CreateMarina(tc.ctx)
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
			req := createTestMarinaRequest("Invalid", orgID)
			req.Name = "" // Name is required
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupMarinaTest(t, http.MethodPost, "/marinas", reqJSON)
			// tc.logger.Info("Testing invalid marina creation")

			// Call handler
			err = tc.handler.CreateMarina(tc.ctx)
			assert.NoError(t, err)

			// Check response for validation error
			assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
		})
	})
}

// TestGetMarinaEndpoints tests various get marina endpoints
func TestGetMarinaEndpoints(t *testing.T) {
	// First create an organization for the marina
	orgID := createOrganizationForTest(t)

	t.Run("Get By ID", func(t *testing.T) {
		// Create a test marina first
		marinaID := createMarinaViaAPI(t, orgID)

		// Test getting by ID
		tc := setupMarinaTest(t, http.MethodGet, "/marinas/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(marinaID.String())
		// tc.logger.Info("Testing get marina by ID", zap.String("id", marinaID.String()))

		// Call handler
		err := tc.handler.GetMarinaByID(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify data exists
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, marinaID.String(), data["id"])
	})

	t.Run("Get By ID Not Found", func(t *testing.T) {
		// Setup test with a non-existent ID
		nonExistentID := uuid.New()
		tc := setupMarinaTest(t, http.MethodGet, "/marinas/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(nonExistentID.String())
		// tc.logger.Info("Testing get marina by non-existent ID", zap.String("id", nonExistentID.String()))

		// Call handler
		err := tc.handler.GetMarinaByID(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusNotFound, tc.recorder.Code)
	})

	t.Run("Get By ID Invalid UUID", func(t *testing.T) {
		// Setup test with an invalid UUID
		tc := setupMarinaTest(t, http.MethodGet, "/marinas/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues("invalid-uuid")
		// tc.logger.Info("Testing get marina with invalid UUID")

		// Call handler
		err := tc.handler.GetMarinaByID(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
	})

	t.Run("Get With Address", func(t *testing.T) {
		// Create a test marina first
		marinaID := createMarinaViaAPI(t, orgID)

		// Test getting with address
		tc := setupMarinaTest(t, http.MethodGet, "/marinas/:id/with-address", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(marinaID.String())
		// tc.logger.Info("Testing get marina with address", zap.String("id", marinaID.String()))

		// Call handler
		err := tc.handler.GetMarinaWithAddress(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify marina and address data exists
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, marinaID.String(), data["id"])

		// Check for address
		address, ok := data["address"].(map[string]interface{})
		assert.True(t, ok, "Address should be included in response")
		assert.NotEmpty(t, address["city"])
	})
}

// TestGetMarinaByEmail tests the get marina by email endpoint
func TestGetMarinaByEmail(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// First create an organization for the marina
		orgID := createOrganizationForTest(t)

		// Create a test marina with a unique email
		randomID := uuid.New().String()
		email := fmt.Sprintf("email-test-marina-%s@example.com", randomID)
		req := createTestMarinaRequest("Email", orgID)
		req.Email = email

		reqJSON, err := json.Marshal(req)
		require.NoError(t, err)

		// Create the marina
		createTC := setupMarinaTest(t, http.MethodPost, "/marinas", reqJSON)
		createTC.logger.Info("Creating test marina for email test", zap.String("email", email))
		err = createTC.handler.CreateMarina(createTC.ctx)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, createTC.recorder.Code)

		// Now test getting the marina by email
		url := fmt.Sprintf("/marinas/by-email?email=%s", email)
		tc := setupMarinaTest(t, http.MethodGet, url, nil)
		tc.ctx.QueryParams().Set("email", email)
		// tc.logger.Info("Testing get marina by email", zap.String("email", email))

		// Call handler
		err = tc.handler.GetMarinaByEmail(tc.ctx)
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

// TestGetMarinasPaginated tests the paginated marinas endpoint
func TestGetMarinasPaginated(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// First create an organization for the marinas
		orgID := createOrganizationForTest(t)

		// Create multiple test marinas
		tc := setupMarinaTest(t, http.MethodGet, "/marinas", nil)
		// tc.logger.Info("Creating multiple test marinas for pagination test")

		// Create 5 marinas
		for i := 0; i < 5; i++ {
			req := createTestMarinaRequest(fmt.Sprintf("Page%d", i), orgID)
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			createTC := setupMarinaTest(t, http.MethodPost, "/marinas", reqJSON)
			err = createTC.handler.CreateMarina(createTC.ctx)
			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, createTC.recorder.Code)
		}

		// Test pagination with limit=3 and offset=0
		url := "/marinas?page=1&pageSize=3"
		tc = setupMarinaTest(t, http.MethodGet, url, nil)
		tc.ctx.QueryParams().Set("page", "1")
		tc.ctx.QueryParams().Set("pageSize", "3")
		// tc.logger.Info("Testing paginated marinas", zap.Int("page", 1), zap.Int("pageSize", 3))

		// Call handler
		err := tc.handler.GetMarinasPaginated(tc.ctx)
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

		// Verify response data is an array of marinas
		data, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Len(t, data, 3)
	})
}

// TestGetMarinasByOrganization tests fetching marinas by organization ID
func TestGetMarinasByOrganization(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// Create an organization for the marinas
		orgID := createOrganizationForTest(t)

		// Create 3 marinas for this organization
		for i := 0; i < 3; i++ {
			req := createTestMarinaRequest(fmt.Sprintf("OrgMarinas%d", i), orgID)
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			createTC := setupMarinaTest(t, http.MethodPost, "/marinas", reqJSON)
			err = createTC.handler.CreateMarina(createTC.ctx)
			require.NoError(t, err)
			require.Equal(t, http.StatusCreated, createTC.recorder.Code)
		}

		// Test getting marinas for this organization
		url := fmt.Sprintf("/marinas/organization/%s?page=1&pageSize=10", orgID.String())
		tc := setupMarinaTest(t, http.MethodGet, url, nil)
		tc.ctx.SetParamNames("organizationId")
		tc.ctx.SetParamValues(orgID.String())
		tc.ctx.QueryParams().Set("page", "1")
		tc.ctx.QueryParams().Set("pageSize", "10")
		// tc.logger.Info("Testing get marinas by organization", zap.String("organizationId", orgID.String()))

		// Call handler
		err := tc.handler.GetMarinasByOrganization(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify response data
		assert.GreaterOrEqual(t, response.Total, int64(3))

		// Verify response data is an array of marinas
		data, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 3)
	})
}

// TestUpdateMarinaEndpoints tests the update marina endpoints
func TestUpdateMarinaEndpoints(t *testing.T) {
	// First create an organization for the marina
	orgID := createOrganizationForTest(t)

	t.Run("Update Marina", func(t *testing.T) {
		// Create a test marina first
		marinaID := createMarinaViaAPI(t, orgID)

		// Create update request
		newCountry := "Spain"
		updateReq := requests.UpdateMarinaRequest{
			Country: &newCountry,
		}

		updateJSON, err := json.Marshal(updateReq)
		require.NoError(t, err)

		// Setup update test
		tc := setupMarinaTest(t, http.MethodPut, "/marinas/:id", updateJSON)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(marinaID.String())
		// tc.logger.Info("Testing update marina", zap.String("id", marinaID.String()))

		// Call update handler
		err = tc.handler.UpdateMarina(tc.ctx)
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

	t.Run("Update Marina With Address", func(t *testing.T) {
		// Create a test marina first
		marinaID := createMarinaViaAPI(t, orgID)

		// Create update request with address changes
		newCity := "Barcelona"
		updateReq := requests.UpdateMarinaRequest{
			Address: &requests.UpdateAddressRequest{
				City: &newCity,
			},
		}

		updateJSON, err := json.Marshal(updateReq)
		require.NoError(t, err)

		// Setup update test
		tc := setupMarinaTest(t, http.MethodPut, "/marinas/:id/with-address", updateJSON)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(marinaID.String())
		// tc.logger.Info("Testing update marina with address", zap.String("id", marinaID.String()))

		// Call update handler
		err = tc.handler.UpdateMarinaWithAddress(tc.ctx)
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

// TestDeleteMarina tests the delete marina endpoint
func TestDeleteMarina(t *testing.T) {
	// First create an organization for the marina
	orgID := createOrganizationForTest(t)

	// Create a test marina first
	marinaID := createMarinaViaAPI(t, orgID)

	// Setup delete test
	tc := setupMarinaTest(t, http.MethodDelete, "/marinas/:id", nil)
	tc.ctx.SetParamNames("id")
	tc.ctx.SetParamValues(marinaID.String())
	// tc.logger.Info("Testing delete marina", zap.String("id", marinaID.String()))

	// Call delete handler
	err := tc.handler.DeleteMarina(tc.ctx)
	assert.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, tc.recorder.Code)

	var response responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify success message
	assert.Equal(t, "Marina successfully deleted", response.Message)

	// Attempt to get the deleted marina (should be soft-deleted)
	getTC := setupMarinaTest(t, http.MethodGet, "/marinas/:id", nil)
	getTC.ctx.SetParamNames("id")
	getTC.ctx.SetParamValues(marinaID.String())
	// getTC.logger.Info("Testing get on deleted marina", zap.String("id", marinaID.String()))

	// This should fail since the marina is soft-deleted
	err = getTC.handler.GetMarinaByID(getTC.ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getTC.recorder.Code)
}
