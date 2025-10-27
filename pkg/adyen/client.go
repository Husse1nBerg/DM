package adyen

import (
	"github.com/adyen/adyen-go-api-library/v14/src/adyen"
	"github.com/adyen/adyen-go-api-library/v14/src/common"
	"github.com/dockworks/dm-web-backend/internal/config"
)

type Client struct {
	APIClient       *adyen.APIClient
	MerchantAccount string
	ClientKey       string
	HMACKey         string
}

func NewClient(cfg *config.AdyenConfig) *Client {
	var environment common.Environment
	switch cfg.Environment {
	case "live":
		environment = common.LiveEnv
	case "test":
		environment = common.TestEnv
	default:
		environment = common.TestEnv
	}

	client := adyen.NewClient(&common.Config{
		ApiKey:      cfg.APIKey,
		Environment: environment,
	})

	return &Client{
		APIClient:       client,
		MerchantAccount: cfg.MerchantAccount,
		ClientKey:       cfg.ClientKey,
		HMACKey:         cfg.HMACKey,
	}
}
