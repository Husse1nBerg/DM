package responses

import "github.com/dockworks/dm-web-backend/pkg/dme"

// VendorListResponse represents the response for listing vendors
type VendorListResponse struct {
	Vendors []VendorDTO `json:"vendors"`
	Count   int         `json:"count"`
}

// VendorSearchResponse represents the response for searching vendors
type VendorSearchResponse struct {
	Vendors []VendorSearchResultDTO `json:"vendors"`
	Count   int                     `json:"count"`
}

// VendorResponse represents the response for a single vendor
type VendorResponse struct {
	Vendor VendorDTO `json:"vendor"`
}

// VendorDTO represents a vendor data transfer object
type VendorDTO struct {
	VendorID             string `json:"vendorId"`
	VendorName           string `json:"vendorName"`
	Address1             string `json:"address1,omitempty"`
	Address2             string `json:"address2,omitempty"`
	Address3             string `json:"address3,omitempty"`
	City                 string `json:"city,omitempty"`
	State                string `json:"state,omitempty"`
	PostalCode           string `json:"postalCode,omitempty"`
	Country              string `json:"country,omitempty"`
	Phone                string `json:"phone,omitempty"`
	Fax                  string `json:"fax,omitempty"`
	ContactName          string `json:"contactName,omitempty"`
	PaymentName          string `json:"paymentName,omitempty"`
	PaymentAddress1      string `json:"paymentAddress1,omitempty"`
	PaymentAddress2      string `json:"paymentAddress2,omitempty"`
	PaymentAddress3      string `json:"paymentAddress3,omitempty"`
	PaymentCity          string `json:"paymentCity,omitempty"`
	PaymentState         string `json:"paymentState,omitempty"`
	PaymentPostalCode    string `json:"paymentPostalCode,omitempty"`
	PaymentCountry       string `json:"paymentCountry,omitempty"`
	PaymentPhone         string `json:"paymentPhone,omitempty"`
	PaymentFax           string `json:"paymentFax,omitempty"`
	PaymentContactName   string `json:"paymentContactName,omitempty"`
	AccountNumber        string `json:"accountNumber,omitempty"`
	EmailAddress         string `json:"emailAddress,omitempty"`
}

// VendorSearchResultDTO represents a simplified vendor search result
type VendorSearchResultDTO struct {
	VendorID      string `json:"vendorId"`
	VendorName    string `json:"vendorName"`
	Phone         string `json:"phone,omitempty"`
	AccountNumber string `json:"accountNumber,omitempty"`
	EmailAddress  string `json:"emailAddress,omitempty"`
	ContactName   string `json:"contactName,omitempty"`
}

// NewVendorListResponse creates a new vendor list response
func NewVendorListResponse(vendors []dme.Vendor) VendorListResponse {
	vendorDTOs := make([]VendorDTO, len(vendors))
	for i, vendor := range vendors {
		vendorDTOs[i] = toVendorDTO(vendor)
	}
	return VendorListResponse{
		Vendors: vendorDTOs,
		Count:   len(vendorDTOs),
	}
}

// NewVendorSearchResponse creates a new vendor search response
func NewVendorSearchResponse(vendors []dme.VendorSearchResult) VendorSearchResponse {
	vendorDTOs := make([]VendorSearchResultDTO, len(vendors))
	for i, vendor := range vendors {
		vendorDTOs[i] = toVendorSearchResultDTO(vendor)
	}
	return VendorSearchResponse{
		Vendors: vendorDTOs,
		Count:   len(vendorDTOs),
	}
}

// NewVendorResponse creates a new single vendor response
func NewVendorResponse(vendor dme.Vendor) VendorResponse {
	return VendorResponse{
		Vendor: toVendorDTO(vendor),
	}
}

// Helper functions to convert DME models to DTOs
func toVendorDTO(vendor dme.Vendor) VendorDTO {
	return VendorDTO{
		VendorID:             vendor.VendorID,
		VendorName:           vendor.VendorName,
		Address1:             vendor.Address1,
		Address2:             vendor.Address2,
		Address3:             vendor.Address3,
		City:                 vendor.City,
		State:                vendor.State,
		PostalCode:           vendor.PostalCode,
		Country:              vendor.Country,
		Phone:                vendor.Phone,
		Fax:                  vendor.Fax,
		ContactName:          vendor.ContactName,
		PaymentName:          vendor.PaymentName,
		PaymentAddress1:      vendor.PaymentAddress1,
		PaymentAddress2:      vendor.PaymentAddress2,
		PaymentAddress3:      vendor.PaymentAddress3,
		PaymentCity:          vendor.PaymentCity,
		PaymentState:         vendor.PaymentState,
		PaymentPostalCode:    vendor.PaymentPostalCode,
		PaymentCountry:       vendor.PaymentCountry,
		PaymentPhone:         vendor.PaymentPhone,
		PaymentFax:           vendor.PaymentFax,
		PaymentContactName:   vendor.PaymentContactName,
		AccountNumber:        vendor.AccountNumber,
		EmailAddress:         vendor.EmailAddress,
	}
}

func toVendorSearchResultDTO(vendor dme.VendorSearchResult) VendorSearchResultDTO {
	return VendorSearchResultDTO{
		VendorID:      vendor.VendorID,
		VendorName:    vendor.VendorName,
		Phone:         vendor.Phone,
		AccountNumber: vendor.AccountNumber,
		EmailAddress:  vendor.EmailAddress,
		ContactName:   vendor.ContactName,
	}
}

