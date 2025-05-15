package requests

import "github.com/google/uuid"

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
	Name                      string `json:"name"`
	FirstName                 string `json:"firstName"`
	LastName                  string `json:"lastName" validate:"required"`
	Address1                  string `json:"address1"`
	Address2                  string `json:"address2"`
	Address3                  string `json:"address3"`
	City                      string `json:"city"`
	State                     string `json:"state"`
	Zip                       string `json:"zip"`
	Country                   string `json:"country"`
	Phone                     string `json:"phone"`
	AltFirstName              string `json:"altFirstName"`
	AltLastName               string `json:"altLastName"`
	AltAddress1               string `json:"altAddress1"`
	AltAddress2               string `json:"altAddress2"`
	AltAddress3               string `json:"altAddress3"`
	AltCity                   string `json:"altCity"`
	AltState                  string `json:"altState"`
	AltZip                    string `json:"altZip"`
	AltCountry                string `json:"altCountry"`
	AltPhone                  string `json:"altPhone"`
	UseAltAddress             bool   `json:"useAltAddress"`
	WorkPhone                 string `json:"workPhone"`
	CellPhone                 string `json:"cellPhone"`
	EmergencyContact          string `json:"emergencyContact"`
	EmergencyPhone            string `json:"emergencyPhone"`
	CompanyName               string `json:"companyName"`
	ShipmentMethod            string `json:"shipmentMethod"`
	ShipmentMethodDescription string `json:"shipmentMethodDescription"`
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
