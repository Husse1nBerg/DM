package adyen

// Payment result codes from Adyen
const (
	PaymentResultAuthorised = "Authorised"
	PaymentResultCaptured   = "Captured"
	PaymentResultRefused    = "Refused"
	PaymentResultError      = "Error"
	PaymentResultCancelled  = "Cancelled"
	PaymentResultReceived   = "Received"
	PaymentResultPending    = "Pending"
)

// Payment event codes from Adyen
const (
	EventAuthorisation  = "AUTHORISATION"
	EventCapture        = "CAPTURE"
	EventCancellation   = "CANCELLATION"
	EventRefund         = "REFUND"
	EventChargeback     = "CHARGEBACK"
	EventRefundReversed = "REFUND_REVERSED"
)
