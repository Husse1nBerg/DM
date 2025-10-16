package responses

import (
	"time"
)

// BatchSubmissionResponse represents the response for batch submission
type BatchSubmissionResponse struct {
	BatchID      string    `json:"batch_id"`
	LocationCode string    `json:"location_code"`
	PostBatch    bool      `json:"post_batch"`
	ReferenceIDs []string  `json:"reference_ids"`
	PostResult   string    `json:"post_result"`
	SubmittedAt  time.Time `json:"submitted_at"`
	TotalAmount  float64   `json:"total_amount"`
	ReceiptCount int       `json:"receipt_count"`
}
