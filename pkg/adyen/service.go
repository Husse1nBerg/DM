package adyen

import (
	"context"
	"fmt"
	"math"

	"github.com/adyen/adyen-go-api-library/v9/src/adyen"
	"github.com/adyen/adyen-go-api-library/v9/src/checkout"
	"github.com/adyen/adyen-go-api-library/v9/src/common"
	"github.com/dockworks/dm-web-backend/internal/config"
	"go.uber.org/zap"
)

type AdyenService struct {
	client *adyen.APIClient
	config *config.AdyenConfig
	logger *zap.Logger
}

func NewAdyenService(cfg *config.AdyenConfig, logger *zap.Logger) (*AdyenService, error) {
	env := common.TestEnv
	if cfg.Environment == "live" {
		env = common.LiveEnv
	}

	client := adyen.NewClient(&common.Config{
		ApiKey:                cfg.APIKey,
		Environment:           env,
		LiveEndpointURLPrefix: cfg.LiveEndpointURL,
	})

	return &AdyenService{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// CreatePaymentSession calls Adyen /sessions using v9 SDK patterns.
func (s *AdyenService) CreatePaymentSession(ctx context.Context, req *checkout.CreateCheckoutSessionRequest) (*checkout.CreateCheckoutSessionResponse, error) {
	s.logger.Info("Creating Adyen payment session", zap.Any("request", req))

	service := s.client.Checkout()

	// Build the SDK input (the generated SDK uses a builder pattern)
	input := service.PaymentsApi.
		SessionsInput().
		CreateCheckoutSessionRequest(*req) // note: takes a value in the builder

	res, httpRes, err := service.PaymentsApi.Sessions(ctx, input)

	statusCode := 0
	if httpRes != nil {
		statusCode = httpRes.StatusCode
	}

	if err != nil {
		s.logger.Error("Failed to create payment session",
			zap.Error(err),
			zap.Int("http_status", statusCode),
		)
		return nil, fmt.Errorf("failed to create payment session: %w", err)
	}

	s.logger.Info("Payment session created",
		zap.String("session_id", res.Id),
		zap.String("merchant_reference", req.Reference),
	)

	return &res, nil
}

// CreateCheckoutSession helper: builds CreateCheckoutSessionRequest correctly
// NOTE: amountFloat is an amount expressed in major units, e.g. 10.50 for €10.50
func (s *AdyenService) CreateCheckoutSession(
	ctx context.Context,
	amountFloat float64,
	currency string,
	returnURL string,
	reference string,
	countryCode string,
	shopperLocale string,
	shopperReference string,
	storeID *string,
) (*checkout.CreateCheckoutSessionResponse, error) {

	value := int64(math.Round(amountFloat * 100)) // convert to minor units safely

	req := &checkout.CreateCheckoutSessionRequest{
		Amount: checkout.Amount{
			Currency: currency,
			Value:    value,
		},
		Reference:        reference,
		ReturnUrl:        returnURL,
		MerchantAccount:  s.config.MerchantAccount,
		CountryCode:      common.PtrString(countryCode),
		ShopperLocale:    common.PtrString(shopperLocale),
		ShopperReference: common.PtrString(shopperReference),
		Store:            storeID,
	}

	return s.CreatePaymentSession(ctx, req)
}

func (s *AdyenService) GetPaymentStatus(ctx context.Context, paymentPspReference string) (*checkout.PaymentDetailsResponse, error) {
	s.logger.Info("Getting payment status", zap.String("psp_reference", paymentPspReference))

	// Use the Checkout Payments API to get payment details
	checkoutAPI := s.client.Checkout()

	// Create payment completion details (empty for status check)
	details := checkout.PaymentCompletionDetails{}

	request := checkout.PaymentDetailsRequest{
		Details:     details,
		PaymentData: &paymentPspReference,
	}

	input := checkoutAPI.PaymentsApi.PaymentsDetailsInput().PaymentDetailsRequest(request)
	response, httpRes, err := checkoutAPI.PaymentsApi.PaymentsDetails(ctx, input)

	statusCode := 0
	if httpRes != nil {
		statusCode = httpRes.StatusCode
	}

	if err != nil {
		s.logger.Error("Failed to get payment status",
			zap.Error(err),
			zap.Int("http_status", statusCode),
			zap.String("psp_reference", paymentPspReference),
		)
		return nil, fmt.Errorf("failed to get payment status: %w", err)
	}

	s.logger.Info("Payment status retrieved successfully",
		zap.String("psp_reference", paymentPspReference),
		zap.String("status", *response.ResultCode),
	)

	return &response, nil
}
