package errors

var (
	PaymentTaxConfigurationNotFound = New("payment tax configuration not found for this marina and payment type")
	PaymentTaxInvalidPaymentType    = New("invalid payment type for this marina")
)
