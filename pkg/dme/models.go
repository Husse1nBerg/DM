package dme

// VersionInfo represents the system version information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
}

// Customer represents customer information
type Customer struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	Email           string  `json:"email"`
	Phone           string  `json:"phone"`
	CompanyName     string  `json:"companyName"`
	BillingAddress  Address `json:"billingAddress"`
	ShippingAddress Address `json:"shippingAddress"`
	Balance         float64 `json:"balance"`
	IsActive        bool    `json:"isActive"`
	CustomerSince   string  `json:"customerSince"`
}

// Prospect represents prospect information
type Prospect struct {
	ID            string  `json:"id"`
	FirstName     string  `json:"firstName"`
	LastName      string  `json:"lastName"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	CompanyName   string  `json:"companyName"`
	Address       Address `json:"address"`
	IsActive      bool    `json:"isActive"`
	ProspectSince string  `json:"prospectSince"`
}

// Address represents a mailing address
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

// Boat represents boat information
type Boat struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Make           string  `json:"make"`
	Model          string  `json:"model"`
	Year           int     `json:"year"`
	Length         float64 `json:"length"`
	RegistrationNo string  `json:"registrationNo"`
	HIN            string  `json:"hin"`
	CustomerId     string  `json:"customerId"`
}

// Invoice represents invoice information
type Invoice struct {
	ID            string  `json:"id"`
	CustomerId    string  `json:"customerId"`
	InvoiceDate   string  `json:"invoiceDate"`
	DueDate       string  `json:"dueDate"`
	Amount        float64 `json:"amount"`
	Balance       float64 `json:"balance"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
	InvoiceNumber string  `json:"invoiceNumber"`
}

// Contract represents contract information
type Contract struct {
	ID             string  `json:"id"`
	CustomerId     string  `json:"customerId"`
	ContractNumber string  `json:"contractNumber"`
	StartDate      string  `json:"startDate"`
	EndDate        string  `json:"endDate"`
	Amount         float64 `json:"amount"`
	Balance        float64 `json:"balance"`
	Status         string  `json:"status"`
	Type           string  `json:"type"`
}

// Reservation represents reservation information
type Reservation struct {
	ID                string `json:"id"`
	CustomerId        string `json:"customerId"`
	BoatId            string `json:"boatId"`
	ReservationNumber string `json:"reservationNumber"`
	ArrivalDate       string `json:"arrivalDate"`
	DepartureDate     string `json:"departureDate"`
	Status            string `json:"status"`
}

// WorkOrder represents a service work order
type WorkOrder struct {
	ID              string `json:"id"`
	CustomerId      string `json:"customerId"`
	BoatId          string `json:"boatId"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	OpenDate        string `json:"openDate"`
	CompleteDate    string `json:"completeDate"`
	WorkOrderNumber string `json:"workOrderNumber"`
	TechId          string `json:"techId"`
}

// PayType represents a payment type
type PayType struct {
	ID                   string `json:"id"`
	Description          string `json:"description"`
	CCSurchargeTaxSchema string `json:"ccSurchargeTaxSchema"`
	IsCheck              bool   `json:"isCheck"`
	IsCreditCard         bool   `json:"isCreditCard"`
	IsActive             bool   `json:"isActive"`
}

// Location represents a location
type Location struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// Clerk represents a clerk
type Clerk struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	IsActive  bool   `json:"isActive"`
}

// CashReceipt represents a cash receipt for batch submission
type CashReceipt struct {
	CustomerId          string              `json:"customerId"`
	ReceiveDate         string              `json:"receiveDate"`
	ReferenceNumber     string              `json:"referenceNumber"`
	PayTypeId           string              `json:"payTypeId"`
	Amount              float64             `json:"amount"`
	ClerkId             string              `json:"clerkId"`
	Description         string              `json:"description"`
	CCTrackData         string              `json:"ccTrackData,omitempty"`
	CCTransactionNumber string              `json:"ccTransactionNumber,omitempty"`
	InvoiceAllocations  []InvoiceAllocation `json:"invoiceAllocations,omitempty"`
}

// InvoiceAllocation represents how a payment should be allocated to invoices
type InvoiceAllocation struct {
	InvoiceId       string  `json:"invoiceId"`
	AllocatedAmount float64 `json:"allocatedAmount"`
}

// MiscARCharge represents a miscellaneous AR charge
type MiscARCharge struct {
	CustomerId        string  `json:"customerId"`
	ChargeDate        string  `json:"chargeDate"`
	ChargeAmount      float64 `json:"chargeAmount"`
	ChargeDescription string  `json:"chargeDescription"`
	LocationCode      string  `json:"locationCode"`
	IsConvenienceFee  bool    `json:"isConvenienceFee"`
}

// ReversePaymentRequest represents a request to reverse a payment
type ReversePaymentRequest struct {
	CustomerId    string `json:"customerId"`
	BatchId       string `json:"batchId"`
	ReferenceId   string `json:"referenceId"`
	StatementDesc string `json:"statementDesc"`
}

// BatchSubmissionRequest represents a request to submit a batch
type BatchSubmissionRequest struct {
	LocationCode string        `json:"locationCode"`
	PostBatch    bool          `json:"postBatch"`
	CashReceipts []CashReceipt `json:"cashReceipts"`
}

// TaxCalculationRequest represents a request to calculate tax
type TaxCalculationRequest struct {
	ReferenceNumber string  `json:"referenceNumber"`
	CustomerId      string  `json:"customerId"`
	LocationCode    string  `json:"locationCode"`
	SaleAmount      float64 `json:"saleAmount"`
	TaxDate         string  `json:"taxDate"`
	TaxSchema       string  `json:"taxSchema"`
}

// CreditMemoApplication represents the application of a credit memo
type CreditMemoApplication struct {
	CreditMemoId    string              `json:"creditMemoId"`
	EffectiveDate   string              `json:"effectiveDate"`
	AutoApply       bool                `json:"autoApply"`
	ApplyToInvoices []CreditMemoInvoice `json:"applyToInvoices"`
}

// CreditMemoInvoice represents an invoice to apply a credit memo to
type CreditMemoInvoice struct {
	InvoiceId     string  `json:"invoiceId"`
	AppliedAmount float64 `json:"appliedAmount"`
}

// WorkOrdersListRequest represents a request to retrieve a list of work orders
type WorkOrdersListRequest struct {
	LastUpdateDate string   `json:"lastUpdateDate"`
	LastUpdateTime string   `json:"lastUpdateTime"`
	TechId         string   `json:"techId"`
	Status         string   `json:"status"`
	WOIds          []string `json:"woIds"`
	Detail         bool     `json:"detail"`
}
