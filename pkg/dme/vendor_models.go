package dme

// Vendor represents a vendor/supplier in the system
type Vendor struct {
	VendorID             string `json:"vendorId"`
	VendorName           string `json:"vendorName"`
	Address1             string `json:"address1"`
	Address2             string `json:"address2"`
	Address3             string `json:"address3"`
	City                 string `json:"city"`
	State                string `json:"state"`
	PostalCode           string `json:"postalCode"`
	Country              string `json:"country"`
	Phone                string `json:"phone"`
	Fax                  string `json:"fax"`
	ContactName          string `json:"contactName"`
	PaymentName          string `json:"paymentName"`
	PaymentAddress1      string `json:"paymentAddress1"`
	PaymentAddress2      string `json:"paymentAddress2"`
	PaymentAddress3      string `json:"paymentAddress3"`
	PaymentCity          string `json:"paymentCity"`
	PaymentState         string `json:"paymentState"`
	PaymentPostalCode    string `json:"paymentPostalCode"`
	PaymentCountry       string `json:"paymentCountry"`
	PaymentPhone         string `json:"paymentPhone"`
	PaymentFax           string `json:"paymentFax"`
	PaymentContactName   string `json:"paymentContactName"`
	AccountNumber        string `json:"accountNumber"`
	EmailAddress         string `json:"emailAddress"`
}

// VendorSearchResult represents a simplified vendor search result
type VendorSearchResult struct {
	VendorID      string `json:"vendorId"`
	VendorName    string `json:"vendorName"`
	Phone         string `json:"phone"`
	AccountNumber string `json:"accountNumber"`
	EmailAddress  string `json:"emailAddress"`
	ContactName   string `json:"contactName"`
}

