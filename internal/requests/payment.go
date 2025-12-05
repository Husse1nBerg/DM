package requests

// InvPayment represents an invoice payment within a cash receipt
type InvPayment struct {
	InvoiceID    string  `json:"invoiceId" validate:"required"`
	LocationCode string  `json:"locationCode" validate:"required"`
	DepositType  string  `json:"depositType" validate:"omitempty"`
	PaymentAmt   float64 `json:"paymentAmt" validate:"required,gt=0"`
	Description  string  `json:"description" validate:"required"`
	CustomerID   string  `json:"customerId" validate:"required"`
}

// CashReceipt represents a single cash receipt in a batch
type CashReceipt struct {
	CustomerID             string       `json:"customerId" validate:"required"`
	ReferenceNum           string       `json:"referenceNum" validate:"required"`
	PayType                string       `json:"payType" validate:"required"`
	TotalPayment           float64      `json:"totalPayment" validate:"required,gt=0"`
	StatementDesc          string       `json:"statementDesc" validate:"required"`
	CCAuthCode             string       `json:"ccAuthCode" validate:"omitempty"`
	CCTransactionID        string       `json:"ccTransactionID" validate:"omitempty"`
	CCTransactionTimeStamp string       `json:"ccTransactionTimeStamp" validate:"omitempty"`
	CCSurcharge            float64      `json:"ccSurcharge" validate:"omitempty"`
	CCSurchargeTax         float64      `json:"ccSurchargeTax" validate:"omitempty"`
	CCSurchargeTaxSchema   string       `json:"ccSurchargeTaxSchema" validate:"omitempty"`
	CCSurchargeTaxIds      []string     `json:"ccSurchargeTaxIds" validate:"omitempty"`
	InvPayments            []InvPayment `json:"invPayments" validate:"required,dive"`
	FirstName              string       `json:"firstName" validate:"omitempty"`
	LastName               string       `json:"lastName" validate:"omitempty"`
	PrimaryEmail           string       `json:"primaryEmail" validate:"omitempty"`
}

// SubmitBatchRequest represents a request to submit a batch of payments
type SubmitBatchRequest struct {
	LocationCode string        `json:"locationCode" validate:"required"`
	PostBatch    bool          `json:"postBatch" validate:"required"`
	CashReceipts []CashReceipt `json:"cashReceipts" validate:"required,dive"`
}
