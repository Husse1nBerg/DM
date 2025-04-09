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

// AddressTestContext holds common test resources
type AddressTestContext struct {
	t        *testing.T
	handler  *handlers.AddressHandler
	recorder *httptest.ResponseRecorder
	ctx      echo.Context
	logger   *zap.SugaredLogger
}

// setupAddressTest creates a new test context with Echo setup
func setupAddressTest(t *testing.T, method, url string, body []byte) *AddressTestContext {
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

	handler := handlers.NewAddressHandler(s)
	return &AddressTestContext{
		t:        t,
		handler:  handler,
		recorder: rec,
		ctx:      ctx,
		logger:   log.Zap,
	}
}

// createTestAddressRequest creates a request for a test address with unique values
func createTestAddressRequest(prefix string) requests.CreateAddressRequest {
	street := fmt.Sprintf("%s Street 123", prefix)
	city := fmt.Sprintf("%s City", prefix)
	state := fmt.Sprintf("%s State", prefix)
	postalCode := "12345"
	country := "USA"
	latitude := 42.123456
	longitude := -71.654321

	return requests.CreateAddressRequest{
		Street:     &street,
		City:       &city,
		State:      &state,
		PostalCode: &postalCode,
		Country:    &country,
		Latitude:   &latitude,
		Longitude:  &longitude,
	}
}

// createAddressViaAPI creates an address via API and returns its ID
func createAddressViaAPI(t *testing.T) uuid.UUID {
	// Create a test address request
	createReq := createTestAddressRequest("Create")
	createJSON, err := json.Marshal(createReq)
	require.NoError(t, err)

	// Create the address using the handler
	tc := setupAddressTest(t, http.MethodPost, "/addresses", createJSON)
	err = tc.handler.CreateAddress(tc.ctx)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, tc.recorder.Code)

	// Extract the address ID from the response
	var createResponse responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &createResponse)
	require.NoError(t, err)

	addressData := createResponse.Data.(map[string]interface{})
	addressIDStr := addressData["id"].(string)
	addressID, err := uuid.Parse(addressIDStr)
	require.NoError(t, err)

	t.Logf("Created test address with ID: %s", addressID)
	return addressID
}

// TestCreateAddress tests the address creation endpoint
func TestCreateAddress(t *testing.T) {
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		t.Run("Valid Address Creation", func(t *testing.T) {
			// Create a valid request with unique values
			req := createTestAddressRequest("Valid")
			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			tc := setupAddressTest(t, http.MethodPost, "/addresses", reqJSON)
			// tc.logger.Info("Testing valid address creation")

			// Call handler
			err = tc.handler.CreateAddress(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusCreated, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify data was created
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, *req.Street, data["street"])
			assert.Equal(t, *req.City, data["city"])
			assert.Equal(t, *req.State, data["state"])
			assert.Equal(t, *req.PostalCode, data["postal_code"])
			assert.Equal(t, *req.Country, data["country"])
			assert.InDelta(t, *req.Latitude, data["latitude"], 0.0001)
			assert.InDelta(t, *req.Longitude, data["longitude"], 0.0001)
		})

		t.Run("Empty Request", func(t *testing.T) {
			// Create an empty request (all fields are optional, so this is valid)
			emptyReq := requests.CreateAddressRequest{}
			reqJSON, err := json.Marshal(emptyReq)
			require.NoError(t, err)

			tc := setupAddressTest(t, http.MethodPost, "/addresses", reqJSON)
			// tc.logger.Info("Testing empty address creation")

			// Call handler
			err = tc.handler.CreateAddress(tc.ctx)
			assert.NoError(t, err)

			// Check response
			assert.Equal(t, http.StatusCreated, tc.recorder.Code)

			var response responses.BaseResponse
			err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify data was created
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.NotEmpty(t, data["id"])
		})
	})
}

// TestGetAddressById tests the get address by ID endpoint
func TestGetAddressById(t *testing.T) {
	t.Run("Get By ID", func(t *testing.T) {
		// Create a test address first
		addressID := createAddressViaAPI(t)

		// Test getting by ID
		tc := setupAddressTest(t, http.MethodGet, "/addresses/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(addressID.String())
		// tc.logger.Info("Testing get address by ID", zap.String("id", addressID.String()))

		// Call handler
		err := tc.handler.GetAddressById(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusOK, tc.recorder.Code)

		var response responses.BaseResponse
		err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify data exists
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, addressID.String(), data["id"])
	})

	t.Run("Get By ID Not Found", func(t *testing.T) {
		// Setup test with a non-existent ID
		nonExistentID := uuid.New()
		tc := setupAddressTest(t, http.MethodGet, "/addresses/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues(nonExistentID.String())
		// tc.logger.Info("Testing get address by non-existent ID", zap.String("id", nonExistentID.String()))

		// Call handler
		err := tc.handler.GetAddressById(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusNotFound, tc.recorder.Code)
	})

	t.Run("Get By ID Invalid UUID", func(t *testing.T) {
		// Setup test with an invalid UUID
		tc := setupAddressTest(t, http.MethodGet, "/addresses/:id", nil)
		tc.ctx.SetParamNames("id")
		tc.ctx.SetParamValues("invalid-uuid")
		// tc.logger.Info("Testing get address with invalid UUID")

		// Call handler
		err := tc.handler.GetAddressById(tc.ctx)
		assert.NoError(t, err)

		// Check response
		assert.Equal(t, http.StatusBadRequest, tc.recorder.Code)
	})
}

// TestUpdateAddress tests the update address endpoint
func TestUpdateAddress(t *testing.T) {
	// Create a test address first
	addressID := createAddressViaAPI(t)

	// Create update request
	newCity := "Updated City"
	newState := "Updated State"
	newLat := 43.210987
	updateReq := requests.UpdateAddressRequest{
		City:     &newCity,
		State:    &newState,
		Latitude: &newLat,
	}

	updateJSON, err := json.Marshal(updateReq)
	require.NoError(t, err)

	// Setup update test
	tc := setupAddressTest(t, http.MethodPut, "/addresses/:id", updateJSON)
	tc.ctx.SetParamNames("id")
	tc.ctx.SetParamValues(addressID.String())
	// tc.logger.Info("Testing update address", zap.String("id", addressID.String()))

	// Call update handler
	err = tc.handler.UpdateAddress(tc.ctx)
	assert.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, tc.recorder.Code)

	var response responses.BaseResponse
	err = json.Unmarshal(tc.recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response data
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, newCity, data["city"])
	assert.Equal(t, newState, data["state"])
	assert.InDelta(t, newLat, data["latitude"], 0.0001)

	// Test update with non-existent address
	nonExistentID := uuid.New()
	tc = setupAddressTest(t, http.MethodPut, "/addresses/:id", updateJSON)
	tc.ctx.SetParamNames("id")
	tc.ctx.SetParamValues(nonExistentID.String())
	// tc.logger.Info("Testing update non-existent address", zap.String("id", nonExistentID.String()))

	// Call update handler
	err = tc.handler.UpdateAddress(tc.ctx)
	assert.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusNotFound, tc.recorder.Code)
}
