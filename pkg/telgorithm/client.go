package telgorithm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/google/uuid"
)

// Client provides an interface to Telgorithm SMS service
type Client struct {
	httpClient  *http.Client
	config      config.TelgorithmConfig
	logger      *logger.Logger
	workerQueue chan smsTask
}

type smsTask struct {
	ID         uuid.UUID
	SMS        *SMSData
	ResultChan chan<- SMSStatus
}

const (
	// Status constants
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"

	// Worker configuration
	workerPoolSize = 5
	queueSize      = 100

	// Endpoints
	sendSMSEndpoint = "/v1/OutboundMessages"

	// Priorities
	PriorityUrgent = "Urgent"
	PriorityHigh   = "High"
	PriorityNormal = "Normal"
	PriorityLow    = "Low"

	// Max concurrent requests for batch operations
	maxConcurrentRequests = 10
)

// NewClient creates a new Telgorithm client
func NewClient(cfg *config.Config) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
		config:      cfg.Telgorithm,
		logger:      &logger.ZLogger,
		workerQueue: make(chan smsTask, queueSize),
	}

	// Start worker pool
	for i := 0; i < workerPoolSize; i++ {
		go client.worker(i)
	}

	return client
}

// worker processes SMS tasks from the queue
func (c *Client) worker(id int) {
	c.logger.Zap.Infow("Starting Telgorithm worker", "worker_id", id)

	for task := range c.workerQueue {
		status := SMSStatus{
			ID:        task.ID,
			Status:    StatusPending,
			Timestamp: time.Now().Unix(),
		}

		// Process multiple recipients one by one as per API doc
		if len(task.SMS.To) == 0 {
			status.Status = StatusFailed
			status.Error = "no recipients specified"

			if task.ResultChan != nil {
				task.ResultChan <- status
				close(task.ResultChan)
			}
			continue
		}

		// Send to each recipient individually
		var lastError error
		var lastResponse *SMSResponse

		for _, recipient := range task.SMS.To {
			resp, err := c.sendSingleSMS(recipient, task.SMS)

			if err != nil {
				lastError = err
				c.logger.Zap.Errorw("Failed to send SMS to recipient",
					"error", err,
					"recipient", recipient,
					"task_id", task.ID.String())
			} else {
				lastResponse = resp
				c.logger.Zap.Infow("SMS sent successfully to recipient",
					"recipient", recipient,
					"task_id", task.ID.String(),
					"message_id", resp.SID)
			}
		}

		// Update the status based on the last operation
		if lastError != nil {
			status.Status = StatusFailed
			status.Error = lastError.Error()
		} else if lastResponse != nil {
			status.Status = StatusSent
			status.MessageID = lastResponse.SID
			status.SegmentCount = lastResponse.SegmentCount
		}

		// Send status on result channel if available
		if task.ResultChan != nil {
			task.ResultChan <- status
			close(task.ResultChan)
		}
	}
}

// SendSMS queues an SMS to be sent asynchronously
func (c *Client) SendSMS(sms *SMSData) (uuid.UUID, <-chan SMSStatus) {
	// Create a new task ID
	taskID := uuid.New()
	resultChan := make(chan SMSStatus, 1)

	// Validate basic requirements
	if len(sms.To) == 0 {
		// If no recipients, return error immediately
		status := SMSStatus{
			ID:        taskID,
			Status:    StatusFailed,
			Error:     "no recipients specified",
			Timestamp: time.Now().Unix(),
		}

		go func() {
			resultChan <- status
			close(resultChan)
		}()

		return taskID, resultChan
	}

	if sms.Message == "" && len(sms.MediaURLs) == 0 {
		// If no message or media, return error immediately
		status := SMSStatus{
			ID:        taskID,
			Status:    StatusFailed,
			Error:     "message or media is required",
			Timestamp: time.Now().Unix(),
		}

		go func() {
			resultChan <- status
			close(resultChan)
		}()

		return taskID, resultChan
	}

	// Queue the task
	task := smsTask{
		ID:         taskID,
		SMS:        sms,
		ResultChan: resultChan,
	}
	c.workerQueue <- task

	return taskID, resultChan
}

// SendBatchSMS sends multiple SMS messages in a batch
// This performs concurrent API calls to the same endpoint
func (c *Client) SendBatchSMS(batch *BatchSMSData) (uuid.UUID, <-chan BatchSMSResult) {
	batchID := uuid.New()
	resultChan := make(chan BatchSMSResult, 1)
	batchResult := BatchSMSResult{
		TaskID:    batchID,
		Results:   make([]SMSStatus, 0, len(batch.Messages)),
		Timestamp: time.Now().Unix(),
	}

	// Validate batch
	if len(batch.Messages) == 0 {
		batchResult.FailureCount = 1
		batchResult.Results = append(batchResult.Results, SMSStatus{
			ID:        batchID,
			Status:    StatusFailed,
			Error:     "batch contains no messages",
			Timestamp: time.Now().Unix(),
		})

		go func() {
			resultChan <- batchResult
			close(resultChan)
		}()

		return batchID, resultChan
	}

	// Process batch asynchronously
	go func() {
		defer close(resultChan)

		// Create a semaphore to limit concurrent requests
		semaphore := make(chan struct{}, maxConcurrentRequests)
		var wg sync.WaitGroup
		var mu sync.Mutex

		for i, message := range batch.Messages {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore

			go func(idx int, sms SMSData) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release semaphore

				// Set default sender if needed
				if sms.FromNumber == "" && batch.FromNumber != "" {
					sms.FromNumber = batch.FromNumber
				}

				// Skip empty messages
				if len(sms.To) == 0 {
					status := SMSStatus{
						ID:        uuid.New(),
						Status:    StatusFailed,
						Error:     "no recipients specified",
						Timestamp: time.Now().Unix(),
					}

					mu.Lock()
					batchResult.Results = append(batchResult.Results, status)
					batchResult.FailureCount++
					mu.Unlock()
					return
				}

				// Process each recipient for this message
				for _, recipient := range sms.To {
					resp, err := c.sendSingleSMS(recipient, &sms)

					status := SMSStatus{
						ID:        uuid.New(),
						Recipient: recipient,
						Timestamp: time.Now().Unix(),
					}

					if err != nil {
						status.Status = StatusFailed
						status.Error = err.Error()

						c.logger.Zap.Errorw("Batch SMS: Failed to send to recipient",
							"batch_id", batchID.String(),
							"message_idx", idx,
							"recipient", recipient,
							"error", err)

						mu.Lock()
						batchResult.FailureCount++
						mu.Unlock()
					} else {
						status.Status = StatusSent
						status.MessageID = resp.SID
						status.SegmentCount = resp.SegmentCount

						c.logger.Zap.Infow("Batch SMS: Successfully sent to recipient",
							"batch_id", batchID.String(),
							"message_idx", idx,
							"recipient", recipient,
							"message_id", resp.SID)

						mu.Lock()
						batchResult.SuccessCount++
						mu.Unlock()
					}

					mu.Lock()
					batchResult.Results = append(batchResult.Results, status)
					mu.Unlock()
				}
			}(i, message)
		}

		// Wait for all requests to complete
		wg.Wait()

		// Send final result
		resultChan <- batchResult
	}()

	return batchID, resultChan
}

// sendSingleSMS sends a single SMS message
func (c *Client) sendSingleSMS(to string, sms *SMSData) (*SMSResponse, error) {
	url := fmt.Sprintf("%s%s", c.config.BaseURL, sendSMSEndpoint)

	// Always use the config FromNumber
	fromNumber := c.config.FromNumber
	if sms.FromNumber != "" {
		fromNumber = sms.FromNumber
	}

	// Prepare request payload
	reqBody := SendSMSRequest{
		From: fromNumber,
		To:   to,
		Text: sms.Message,
	}
	c.logger.Zap.Infow("Sending SMS", "request", reqBody)

	// Add optional fields if provided
	if sms.Priority != "" {
		reqBody.Priority = sms.Priority
	}

	if len(sms.MediaURLs) > 0 {
		reqBody.MediaURLs = sms.MediaURLs
	}

	if sms.ExpiresOn != nil {
		reqBody.ExpiresOn = sms.ExpiresOn.Format(time.RFC3339)
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.config.Username, c.config.Password)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful status code
	if resp.StatusCode != http.StatusOK {
		c.logger.Zap.Errorw("Failed to send SMS",
			"status_code", resp.StatusCode,
			"error", resp.Body)
		return nil, fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	// Parse response
	var smsResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &smsResp, nil
}
