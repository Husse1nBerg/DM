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

// PaymentInitiationResponse represents the response when initiating a payment
type PaymentInitiationResponse struct {
	Data dme.PaymentInitiationResponse `json:"data"`
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

// ConvertPaymentInitiation converts DME PaymentInitiationResponse to PaymentInitiationResponse
func ConvertPaymentInitiation(payment *dme.PaymentInitiationResponse) *PaymentInitiationResponse {
	return &PaymentInitiationResponse{
		Data: *payment,
	}
}

// swag:response InvoiceListResponse
type SwagInvoiceListResponse = InvoiceListResponse

// swag:response PaymentInitiationResponse
type SwagPaymentInitiationResponse = PaymentInitiationResponse
