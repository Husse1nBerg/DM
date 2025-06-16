package requests

import "github.com/dockworks/dm-web-backend/pkg/dme"

// BoatsListRequest represents a request to list boats by page
type BoatsListRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// CustomerRetrieveRequest represents a request to retrieve a customer by ID
type BoatRetrieveRequest struct {
	BoatID string `query:"BoatId" validate:"required"`
}

// BoatSearchRequest represents a request to search for boats
type BoatSearchRequest struct {
	SearchString string `query:"SearchString" validate:"required"`
	DirectHit    bool   `query:"DirectHit"`
}

// BoatCreateRequest represents a request to create a new boat
// It mirrors dme.BoatCreate but is used for input validation
// No id field should be present
type BoatCreateRequest struct {
	OwnerID              string                    `json:"ownerId" validate:"required"`
	Name                 string                    `json:"name" validate:"required"`
	Registration         string                    `json:"registration"`
	Year                 string                    `json:"year"`
	Make                 string                    `json:"make"`
	Model                string                    `json:"model"`
	Hin                  string                    `json:"hin"`
	LOA                  string                    `json:"loa"`
	LWL                  string                    `json:"lwl"`
	Draft                string                    `json:"draft"`
	Beam                 string                    `json:"beam"`
	Height               string                    `json:"height"`
	Color                string                    `json:"color"`
	TrailerMake          string                    `json:"trailerMake"`
	TrailerModel         string                    `json:"trailerModel"`
	TrailerSerial        string                    `json:"trailerSerial"`
	TrailerRegistration  string                    `json:"trailerRegistration"`
	TrailerLocation      string                    `json:"trailerLocation"`
	SummerSlip           string                    `json:"summerSlip"`
	WinterSlip           string                    `json:"winterSlip"`
	InsuranceCompany     string                    `json:"insuranceCompany"`
	InsuranceExpDate     string                    `json:"insuranceExpDate"`
	SlipID               string                    `json:"slipId"`
	DoNotLaunch          bool                      `json:"doNotLaunch"`
	BillingCodes         []dme.BillingCode         `json:"billingCodes"`
	BoatDescriptionCodes []dme.BoatDescriptionCode `json:"boatDescriptionCodes"`
	CustomInformation    []dme.CustomInformation   `json:"customInformation"`
	OperationsHistory    []dme.OperationHistory    `json:"operationsHistory"`
	IntegrationID        string                    `json:"integrationId"`
	OwnerIntegrationID   string                    `json:"ownerIntegrationId"`
	LastModified         string                    `json:"lastModified"`
	Comments             string                    `json:"comments"`
	Attachments          []dme.Attachment          `json:"attachments"`
}

// BoatUpdateRequest represents a request to update an existing boat
type BoatUpdateRequest struct {
	ID                   string                    `json:"id" validate:"required"`
	Name                 string                    `json:"name"`
	Registration         string                    `json:"registration"`
	Year                 string                    `json:"year"`
	Make                 string                    `json:"make"`
	Model                string                    `json:"model"`
	HIN                  string                    `json:"hin"`
	LOA                  string                    `json:"loa"`
	LWL                  string                    `json:"lwl"`
	Draft                string                    `json:"draft"`
	Beam                 string                    `json:"beam"`
	Height               string                    `json:"height"`
	Color                string                    `json:"color"`
	TrailerMake          string                    `json:"trailerMake"`
	TrailerModel         string                    `json:"trailerModel"`
	TrailerSerial        string                    `json:"trailerSerial"`
	TrailerRegistration  string                    `json:"trailerRegistration"`
	TrailerLocation      string                    `json:"trailerLocation"`
	SummerSlip           string                    `json:"summerSlip"`
	WinterSlip           string                    `json:"winterSlip"`
	InsuranceCompany     string                    `json:"insuranceCompany"`
	InsuranceExpDate     string                    `json:"insuranceExpDate"`
	SlipID               string                    `json:"slipId"`
	DoNotLaunch          bool                      `json:"doNotLaunch"`
	BillingCodes         []dme.BillingCode         `json:"billingCodes"`
	BoatDescriptionCodes []dme.BoatDescriptionCode `json:"boatDescriptionCodes"`
	CustomInformation    []dme.CustomInformation   `json:"customInformation"`
	OperationsHistory    []dme.OperationHistory    `json:"operationsHistory"`
	IntegrationID        string                    `json:"integrationId"`
	OwnerIntegrationID   string                    `json:"ownerIntegrationId"`
	LastModified         string                    `json:"lastModified"`
	Comments             string                    `json:"comments"`
	Attachments          []dme.Attachment          `json:"attachments"`
}
