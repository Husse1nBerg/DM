package requests

// GetCustomerInvoicesRequest represents a request to get invoices for a customer
type GetCustomerInvoicesRequest struct {
	CustomerID  string `query:"customerId" validate:"required"`
	InvoiceDate string `query:"invoiceDate"`
}
