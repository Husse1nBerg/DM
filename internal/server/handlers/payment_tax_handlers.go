package handlers

import (
	"fmt"
	"math"
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PaymentTaxHandler struct {
	server *s.Server
}

func NewPaymentTaxHandler(server *s.Server) *PaymentTaxHandler {
	return &PaymentTaxHandler{server: server}
}

// CreatePaymentTax creates a new payment tax configuration
//
//	@Summary		Create payment tax configuration
//	@Description	Creates a new payment tax configuration for convenience fees and surcharges for a marina
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreatePaymentTaxRequest	true	"Payment tax configuration request"
//	@Success		201		{object}	responses.PaymentTaxResponse		"Payment tax configuration created successfully"
//	@Failure		400		{object}	responses.Error						"Invalid request"
//	@Failure		500		{object}	responses.Error						"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax [post]
func (h *PaymentTaxHandler) CreatePaymentTax(c echo.Context) error {
	var req requests.CreatePaymentTaxRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Failed to bind payment tax request", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Payment tax request validation failed", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed").JSON(c)
	}

	// Get user ID from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	// Verify marina exists
	_, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), req.MarinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Marina not found", zap.String("marinaId", req.MarinaID.String()), zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
	}

	// Convert float64 to pgtype.Numeric for convenience fee
	convenienceFee := pgtype.Numeric{}
	if err := convenienceFee.Scan(fmt.Sprintf("%.2f", req.ConvenienceFee)); err != nil {
		h.server.Logger.Zap.Error("Failed to convert convenience fee", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid convenience fee value").JSON(c)
	}

	// Convert float64 to pgtype.Numeric for surcharge
	surcharge := pgtype.Numeric{}
	if err := surcharge.Scan(fmt.Sprintf("%.2f", req.Surcharge)); err != nil {
		h.server.Logger.Zap.Error("Failed to convert surcharge", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid surcharge value").JSON(c)
	}

	// Set default values for optional fields
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create the payment tax configuration
	config, err := h.server.DB.Queries().CreateTaxConfiguration(c.Request().Context(), db.CreateTaxConfigurationParams{
		MarinaID:                  req.MarinaID,
		ConvenienceFee:            convenienceFee,
		ConvenienceFeeType:        req.ConvenienceFeeType,
		ConvenienceFeeEnabled:     req.ConvenienceFeeEnabled,
		ConvenienceFeeDescription: req.ConvenienceFeeDescription,
		Surcharge:                 surcharge,
		SurchargeType:             req.SurchargeType,
		SurchargeEnabled:          req.SurchargeEnabled,
		SurchargeDescription:      req.SurchargeDescription,
		TaxRate:                   req.TaxRate,
		TaxEnabled:                req.TaxEnabled,
		TaxDescription:            req.TaxDescription,
		IsActive:                  isActive,
		CreatedBy:                 userID,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Failed to create payment tax configuration", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create payment tax configuration").JSON(c)
	}

	response := responses.ConvertPaymentTaxToResponse(config)
	return c.JSON(http.StatusCreated, responses.NewSuccessResponse(response))
}

// GetPaymentTax retrieves a payment tax configuration by ID
//
//	@Summary		Get payment tax configuration
//	@Description	Retrieves a payment tax configuration by ID
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string						true	"Payment tax configuration ID"
//	@Success		200	{object}	responses.PaymentTaxResponse	"Payment tax configuration"
//	@Failure		400	{object}	responses.Error					"Invalid request"
//	@Failure		404	{object}	responses.Error					"Not found"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax/{id} [get]
func (h *PaymentTaxHandler) GetPaymentTax(c echo.Context) error {
	var req requests.GetPaymentTaxRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed").JSON(c)
	}

	configID, err := uuid.Parse(req.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	config, err := h.server.DB.Queries().GetTaxConfigurationByID(c.Request().Context(), configID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get payment tax configuration", zap.String("id", req.ID), zap.Error(err))
		return responses.NewErrorResponse(http.StatusNotFound, "Payment tax configuration not found").JSON(c)
	}

	response := responses.ConvertPaymentTaxToResponse(config)
	return c.JSON(http.StatusOK, responses.NewSuccessResponse(response))
}

// GetActiveMarinaPaymentTax retrieves the active payment tax configuration for a marina
//
//	@Summary		Get active marina payment tax configuration
//	@Description	Retrieves the active payment tax configuration for the current user's marina
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.PaymentTaxResponse	"Active payment tax configuration"
//	@Failure		404	{object}	responses.Error					"Not found"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax/active [get]
func (h *PaymentTaxHandler) GetActiveMarinaPaymentTax(c echo.Context) error {
	// Get marina ID from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaID := claims.MarinaId

	config, err := h.server.DB.Queries().GetActiveTaxConfigurationByMarinaID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get active payment tax configuration", zap.String("marinaId", marinaID.String()), zap.Error(err))
		return responses.NewErrorResponse(http.StatusNotFound, "Active payment tax configuration not found").JSON(c)
	}

	response := responses.ConvertPaymentTaxToResponse(config)
	return c.JSON(http.StatusOK, responses.NewSuccessResponse(response))
}

// ListPaymentTax lists payment tax configurations for the current marina
//
//	@Summary		List payment tax configurations
//	@Description	Lists all payment tax configurations for the current user's marina with pagination
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int							false	"Page number"	default(1) minimum(1)
//	@Param			pageSize	query		int							false	"Page size"		default(10) minimum(1) maximum(100)
//	@Success		200			{object}	responses.PaymentTaxListResponse	"Paginated list of payment tax configurations"
//	@Failure		500			{object}	responses.Error					"Internal server error"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax [get]
func (h *PaymentTaxHandler) ListPaymentTax(c echo.Context) error {
	var req requests.ListPaymentTaxRequest
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// Get marina ID from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaID := claims.MarinaId

	configs, err := h.server.DB.Queries().GetAllTaxConfigurationsByMarinaID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to list payment tax configurations", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to list payment tax configurations").JSON(c)
	}

	// Convert to responses
	var responseData []responses.PaymentTaxResponse
	for _, config := range configs {
		responseData = append(responseData, responses.ConvertPaymentTaxToResponse(config))
	}

	// Calculate pagination
	total := int64(len(configs))
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	// Apply pagination to data
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start > len(responseData) {
		start = len(responseData)
	}
	if end > len(responseData) {
		end = len(responseData)
	}
	paginatedData := responseData[start:end]

	response := responses.PaymentTaxListResponse{
		Data:       paginatedData,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	return c.JSON(http.StatusOK, responses.NewSuccessResponse(response))
}

// UpdatePaymentTax updates a payment tax configuration
//
//	@Summary		Update payment tax configuration
//	@Description	Updates an existing payment tax configuration
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Payment tax configuration ID"
//	@Param			request	body		requests.UpdatePaymentTaxRequest	true	"Update request"
//	@Success		200		{object}	responses.PaymentTaxResponse		"Updated payment tax configuration"
//	@Failure		400		{object}	responses.Error						"Invalid request"
//	@Failure		404		{object}	responses.Error						"Not found"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax/{id} [put]
func (h *PaymentTaxHandler) UpdatePaymentTax(c echo.Context) error {
	configID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	var req requests.UpdatePaymentTaxRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Failed to bind update request", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Update request validation failed", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed").JSON(c)
	}

	// Check if config exists
	existingConfig, err := h.server.DB.Queries().GetTaxConfigurationByID(c.Request().Context(), configID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Payment tax configuration not found").JSON(c)
	}

	// Start with default values from existing config
	convenienceFee := 0.0
	if req.ConvenienceFee != nil {
		convenienceFee = *req.ConvenienceFee
	} else {
		convenienceFee, _ = utils.NumericToFloat64(existingConfig.ConvenienceFee)
	}

	surcharge := 0.0
	if req.Surcharge != nil {
		surcharge = *req.Surcharge
	} else {
		surcharge, _ = utils.NumericToFloat64(existingConfig.Surcharge)
	}

	taxRate := existingConfig.TaxRate
	if req.TaxRate != nil {
		taxRate = *req.TaxRate
	}

	// Update the configuration
	config, err := h.server.DB.Queries().UpdateTaxConfiguration(c.Request().Context(), db.UpdateTaxConfigurationParams{
		ID:                        configID,
		ConvenienceFee:            convenienceFee,
		ConvenienceFeeType:        req.ConvenienceFeeType,
		ConvenienceFeeEnabled:     req.ConvenienceFeeEnabled,
		ConvenienceFeeDescription: req.ConvenienceFeeDescription,
		Surcharge:                 surcharge,
		SurchargeType:             req.SurchargeType,
		SurchargeEnabled:          req.SurchargeEnabled,
		SurchargeDescription:      req.SurchargeDescription,
		TaxRate:                   taxRate,
		TaxEnabled:                req.TaxEnabled,
		TaxDescription:            req.TaxDescription,
		IsActive:                  req.IsActive,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Failed to update payment tax configuration", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to update payment tax configuration").JSON(c)
	}

	response := responses.ConvertPaymentTaxToResponse(config)
	return c.JSON(http.StatusOK, responses.NewSuccessResponse(response))
}

// DeactivatePaymentTax deactivates a payment tax configuration
//
//	@Summary		Deactivate payment tax configuration
//	@Description	Deactivates a payment tax configuration
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string				true	"Payment tax configuration ID"
//	@Success		200	{object}	responses.BaseResponse	"Configuration deactivated successfully"
//	@Failure		400	{object}	responses.Error		"Invalid request"
//	@Failure		404	{object}	responses.Error		"Not found"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax/{id}/deactivate [post]
func (h *PaymentTaxHandler) DeactivatePaymentTax(c echo.Context) error {
	configID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	// Check if config exists
	_, err = h.server.DB.Queries().GetTaxConfigurationByID(c.Request().Context(), configID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Payment tax configuration not found").JSON(c)
	}

	err = h.server.DB.Queries().DeactivateTaxConfiguration(c.Request().Context(), configID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to deactivate payment tax configuration", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to deactivate configuration").JSON(c)
	}

	return c.JSON(http.StatusOK, responses.NewSuccessResponse("Configuration deactivated successfully"))
}

// CalculateFees calculates all fees for a given amount using the active configuration
//
//	@Summary		Calculate fees
//	@Description	Calculates convenience fee, surcharge, and tax for a given amount using the active configuration
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CalculateFeeRequest		true	"Amount to calculate fees for"
//	@Success		200		{object}	responses.FeeCalculationResponse	"Calculated fees"
//	@Failure		400		{object}	responses.Error						"Invalid request"
//	@Failure		404		{object}	responses.Error						"No active configuration found"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax/calculate [post]
func (h *PaymentTaxHandler) CalculateFees(c echo.Context) error {
	var req requests.CalculateFeeRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed").JSON(c)
	}

	// Get marina ID from JWT token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaID := claims.MarinaId

	// Get active configuration
	config, err := h.server.DB.Queries().GetActiveTaxConfigurationByMarinaID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get active configuration", zap.Error(err))
		return responses.NewErrorResponse(http.StatusNotFound, "No active payment tax configuration found").JSON(c)
	}

	// Calculate convenience fee
	var convenienceFee float64
	if config.ConvenienceFeeEnabled {
		convenienceFeeValue, _ := utils.NumericToFloat64(config.ConvenienceFee)
		if config.ConvenienceFeeType == "percentage" {
			convenienceFee = req.Amount * convenienceFeeValue / 100.0
		} else {
			convenienceFee = convenienceFeeValue
		}
	}

	// Calculate surcharge
	var surchargeAmount float64
	if config.SurchargeEnabled {
		surchargeValue, _ := utils.NumericToFloat64(config.Surcharge)
		if config.SurchargeType == "percentage" {
			surchargeAmount = req.Amount * surchargeValue / 100.0
		} else {
			surchargeAmount = surchargeValue
		}
	}

	// Calculate tax
	var taxAmount float64
	if config.TaxEnabled {
		taxAmount = req.Amount * config.TaxRate / 100.0
	}

	// Calculate total
	totalAmount := req.Amount + convenienceFee + surchargeAmount + taxAmount

	convenienceFeeRate, _ := utils.NumericToFloat64(config.ConvenienceFee)
	surchargeRate, _ := utils.NumericToFloat64(config.Surcharge)

	response := responses.FeeCalculationResponse{
		BaseAmount:         req.Amount,
		ConvenienceFee:     convenienceFee,
		Surcharge:          surchargeAmount,
		Tax:                taxAmount,
		TotalAmount:        totalAmount,
		ConvenienceFeeType: config.ConvenienceFeeType,
		SurchargeType:      config.SurchargeType,
		ConvenienceFeeRate: convenienceFeeRate,
		SurchargeRate:      surchargeRate,
		TaxRate:            config.TaxRate,
	}

	return c.JSON(http.StatusOK, responses.NewSuccessResponse(response))
}

// DeletePaymentTax deletes a payment tax configuration
//
//	@Summary		Delete payment tax configuration
//	@Description	Permanently deletes a payment tax configuration
//	@Tags			PaymentTax
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string				true	"Payment tax configuration ID"
//	@Success		200	{object}	responses.BaseResponse	"Configuration deleted successfully"
//	@Failure		400	{object}	responses.Error		"Invalid request"
//	@Failure		404	{object}	responses.Error		"Not found"
//	@Security		ApiKeyAuth
//	@Router			/payment-tax/{id} [delete]
func (h *PaymentTaxHandler) DeletePaymentTax(c echo.Context) error {
	configID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	// Check if config exists
	_, err = h.server.DB.Queries().GetTaxConfigurationByID(c.Request().Context(), configID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Payment tax configuration not found").JSON(c)
	}

	err = h.server.DB.Queries().DeleteTaxConfiguration(c.Request().Context(), configID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to delete payment tax configuration", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to delete configuration").JSON(c)
	}

	return c.JSON(http.StatusOK, responses.NewSuccessResponse("Configuration deleted successfully"))
}
