package responses

import "github.com/dockworks/dm-web-backend/pkg/dme"

// InvoiceListResponse represents a list of invoices
type InvoiceListResponse struct {
	Data []dme.InvoiceDetailed `json:"data"`
}

// CustomerInvoiceInquiryListResponse represents a list of customer invoice inquiries
type CustomerInvoiceInquiryListResponse struct {
	Data []dme.CustomerInvoiceInquiry `json:"data"`
}

// ConvertInvoiceList converts a slice of DME InvoiceDetailed to InvoiceListResponse
func ConvertInvoiceList(invoices []dme.InvoiceDetailed) *InvoiceListResponse {
	return &InvoiceListResponse{
		Data: invoices,
	}
}

// ConvertCustomerInvoiceInquiryList converts a slice of DME CustomerInvoiceInquiry to InvoiceListResponse
func ConvertCustomerInvoiceInquiryList(invoices []dme.CustomerInvoiceInquiry) *CustomerInvoiceInquiryListResponse {
	return &CustomerInvoiceInquiryListResponse{
		Data: invoices,
	}
}

// swag:response InvoiceListResponse
type SwagInvoiceListResponse = InvoiceListResponse

// NextReferenceResponse represents the next AR reference number
type NextReferenceResponse struct {
	ReferenceNumber string `json:"referenceNumber"`
}
