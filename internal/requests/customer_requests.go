package requests

import (
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/google/uuid"
)

// CustomerListRequest represents a request to list customers by page
type CustomerListRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// CustomerRetrieveRequest represents a request to retrieve a customer by ID
type CustomerRetrieveRequest struct {
	CustomerID string `query:"CustomerId" validate:"required"`
}

// CustomerSearchRequest represents a request to search for customers
type CustomerSearchRequest struct {
	SearchString string `query:"SearchString" validate:"required"`
	DirectHit    string `query:"DirectHit" validate:"oneof=true false"`
}

// CustomerCreateRequest represents a request to create a new customer
// It mirrors dme.CustomerCreate but is used for input validation
// No id field should be present
type CustomerCreateRequest struct {
	Name                      string                  `json:"name"`
	FirstName                 string                  `json:"firstName"`
	LastName                  string                  `json:"lastName" validate:"required"`
	Email                     string                  `json:"email"`
	Address1                  string                  `json:"address1"`
	Address2                  string                  `json:"address2"`
	Address3                  string                  `json:"address3"`
	City                      string                  `json:"city"`
	State                     string                  `json:"state"`
	Zip                       string                  `json:"zip"`
	Country                   string                  `json:"country"`
	Phone                     string                  `json:"phone"`
	AltFirstName              string                  `json:"altFirstName"`
	AltLastName               string                  `json:"altLastName"`
	AltAddress1               string                  `json:"altAddress1"`
	AltAddress2               string                  `json:"altAddress2"`
	AltAddress3               string                  `json:"altAddress3"`
	AltCity                   string                  `json:"altCity"`
	AltState                  string                  `json:"altState"`
	AltZip                    string                  `json:"altZip"`
	AltCountry                string                  `json:"altCountry"`
	AltPhone                  string                  `json:"altPhone"`
	UseAltAddress             bool                    `json:"useAltAddress"`
	WorkPhone                 string                  `json:"workPhone"`
	CellPhone                 string                  `json:"cellPhone"`
	EmergencyContact          string                  `json:"emergencyContact"`
	EmergencyPhone            string                  `json:"emergencyPhone"`
	CompanyName               string                  `json:"companyName"`
	ShipmentMethod            string                  `json:"shipmentMethod"`
	ShipmentMethodDescription string                  `json:"shipmentMethodDescription"`
	CustomInformation         []dme.CustomInformation `json:"customInformation"`
	Attachments               []dme.Attachment        `json:"attachments"`
}

// CustomerSettingsRetrieveRequest represents a request to retrieve customer settings
type CustomerSettingsRetrieveRequest struct {
	MarinaID   uuid.UUID `query:"MarinaId" validate:"required"`
	CustomerID string    `query:"CustomerId" validate:"required"`
}

// CustomerSettingsUpdateRequest represents a request to update customer settings
type CustomerSettingsUpdateRequest struct {
	MarinaID     uuid.UUID `json:"marinaId" validate:"required"`
	CustomerID   string    `json:"customerId" validate:"required"`
	EnablePortal *bool     `json:"enablePortal"`
}

// CustomerIntakeRequest represents a request to create both a customer and user in one operation
type CustomerIntakeRequest struct {
	// Customer fields
	OrganizationID   string `json:"organizationId" validate:"required"`
	MarinaID         string `json:"marinaId" validate:"required"`
	FirstName        string `json:"firstName" validate:"required"`
	LastName         string `json:"lastName" validate:"required"`
	Email            string `json:"email" validate:"required,email"`
	Address1         string `json:"address1"`
	Address2         string `json:"address2"`
	Address3         string `json:"address3"`
	City             string `json:"city"`
	State            string `json:"state"`
	Zip              string `json:"zip"`
	Country          string `json:"country"`
	Phone            string `json:"phone"`
	WorkPhone        string `json:"workPhone"`
	CellPhone        string `json:"cellPhone"`
	EmergencyContact string `json:"emergencyContact"`
	EmergencyPhone   string `json:"emergencyPhone"`
	CompanyName      string `json:"companyName"`
	Password         string `json:"password"`
}

// AttachmentWithPublic extends dme.Attachment with a Public field for update requests
type AttachmentWithPublic struct {
	dme.Attachment
	Public bool `json:"public"`
}

// CustomerUpdateRequest represents a request to update an existing customer
// All fields except ID use pointers to support partial updates:
// - nil means don't update this field
// - pointer with value means update to that value (even if empty string)
type CustomerUpdateRequest struct {
	ID                        string                  `json:"id" validate:"required"`
	Name                      *string                 `json:"name"`
	FirstName                 *string                 `json:"firstName"`
	LastName                  *string                 `json:"lastName"`
	Email                     *string                 `json:"email"`
	Address1                  *string                 `json:"address1"`
	Address2                  *string                 `json:"address2"`
	Address3                  *string                 `json:"address3"`
	City                      *string                 `json:"city"`
	State                     *string                 `json:"state"`
	Zip                       *string                 `json:"zip"`
	Country                   *string                 `json:"country"`
	Phone                     *string                 `json:"phone"`
	AltFirstName              *string                 `json:"altFirstName"`
	AltLastName               *string                 `json:"altLastName"`
	AltAddress1               *string                 `json:"altAddress1"`
	AltAddress2               *string                 `json:"altAddress2"`
	AltAddress3               *string                 `json:"altAddress3"`
	AltCity                   *string                 `json:"altCity"`
	AltState                  *string                 `json:"altState"`
	AltZip                    *string                 `json:"altZip"`
	AltCountry                *string                 `json:"altCountry"`
	AltPhone                  *string                 `json:"altPhone"`
	UseAltAddress             *bool                   `json:"useAltAddress"`
	WorkPhone                 *string                 `json:"workPhone"`
	CellPhone                 *string                 `json:"cellPhone"`
	EmergencyContact          *string                 `json:"emergencyContact"`
	EmergencyPhone            *string                 `json:"emergencyPhone"`
	CompanyName               *string                 `json:"companyName"`
	ShipmentMethod            *string                 `json:"shipmentMethod"`
	ShipmentMethodDescription *string                 `json:"shipmentMethodDescription"`
	CustomInformation         []dme.CustomInformation `json:"customInformation"`
	Attachments               []AttachmentWithPublic  `json:"attachments"`
	Inactive                  *bool                   `json:"inactive"`
}

// CustomerRetrieveListRequest represents a request to retrieve multiple customers with filters
type CustomerRetrieveListRequest struct {
	LastModifiedDate string `query:"LastModifiedDate"` // Optional: Format MM-DD-YYYY
	EmailAddress     string `query:"EmailAddress"`     // Optional: Retrieves records with matching primary email
}

// CustomerRetrievePaginatedRequest represents a request to retrieve customers with category codes in a paginated format
type CustomerRetrievePaginatedRequest struct {
	Page             int    `query:"Page" validate:"required,min=1"`        // Current page (1-based), required
	PageSize         int    `query:"PageSize" validate:"required,min=1,max=500"` // Items per page, required
	ListName         string `query:"ListName"`                              // Optional: Cached list name for subsequent pages
	LastModifiedDate string `query:"LastModifiedDate"`                      // Optional: Format MM-DD-YYYY
	EmailAddress     string `query:"EmailAddress"`                          // Optional: Retrieves records with matching primary email
}