package sendgrid

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Client provides an interface to SendGrid email service
type Client struct {
	client      *sendgrid.Client
	config      config.SendGridConfig
	logger      *logger.Logger
	workerQueue chan emailTask
}

type emailTask struct {
	ID         uuid.UUID
	TaskType   string // "html" or "template"
	HTMLEmail  *HTMLEmail
	TmplEmail  *TemplateEmail
	ResultChan chan<- EmailStatus
}

const (
	// Status constants
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"

	// Task types
	TaskTypeHTML     = "html"
	TaskTypeTemplate = "template"

	// Worker configuration
	workerPoolSize = 5
	queueSize      = 100
)

// NewClient creates a new SendGrid client
func NewClient(cfg *config.Config) *Client {
	client := sendgrid.NewSendClient(cfg.SendGrid.APIKey)

	sgClient := &Client{
		client:      client,
		config:      cfg.SendGrid,
		logger:      &logger.ZLogger,
		workerQueue: make(chan emailTask, queueSize),
	}

	// Start worker pool
	for i := 0; i < workerPoolSize; i++ {
		go sgClient.worker(i)
	}

	return sgClient
}

// worker processes email tasks from the queue
func (c *Client) worker(id int) {
	c.logger.Zap.Infow("Starting SendGrid worker", "worker_id", id)

	for task := range c.workerQueue {
		status := EmailStatus{
			ID:        task.ID,
			Status:    StatusPending,
			Timestamp: time.Now().Unix(),
		}

		var err error
		if task.TaskType == TaskTypeHTML && task.HTMLEmail != nil {
			err = c.sendHTMLEmailSync(task.HTMLEmail)
		} else if task.TaskType == TaskTypeTemplate && task.TmplEmail != nil {
			err = c.sendTemplateEmailSync(task.TmplEmail)
		} else {
			err = errors.New("invalid task type or missing email data")
		}

		if err != nil {
			status.Status = StatusFailed
			status.Error = err.Error()
			c.logger.Zap.Errorw("Failed to send email",
				"error", err,
				"task_id", task.ID.String(),
				"task_type", task.TaskType)
		} else {
			status.Status = StatusSent
			c.logger.Zap.Infow("Email sent successfully",
				"task_id", task.ID.String(),
				"task_type", task.TaskType)
		}

		// Send status on result channel if available
		if task.ResultChan != nil {
			task.ResultChan <- status
			close(task.ResultChan)
		}
	}
}

// SendHTMLEmail queues an HTML email to be sent asynchronously
func (c *Client) SendHTMLEmail(email *HTMLEmail) (uuid.UUID, <-chan EmailStatus) {
	// Set default sender if not provided
	if email.FromEmail == "" {
		email.FromEmail = c.config.FromEmail
	}
	if email.FromName == "" {
		email.FromName = c.config.FromName
	}

	taskID := uuid.New()
	resultChan := make(chan EmailStatus, 1)

	task := emailTask{
		ID:         taskID,
		TaskType:   TaskTypeHTML,
		HTMLEmail:  email,
		ResultChan: resultChan,
	}

	// Queue the task
	c.workerQueue <- task
	return taskID, resultChan
}

// SendTemplateEmail queues a template email to be sent asynchronously
func (c *Client) SendTemplateEmail(email *TemplateEmail) (uuid.UUID, <-chan EmailStatus) {
	// Set default sender if not provided
	if email.FromEmail == "" {
		email.FromEmail = c.config.FromEmail
	}
	if email.FromName == "" {
		email.FromName = c.config.FromName
	}

	taskID := uuid.New()
	resultChan := make(chan EmailStatus, 1)

	task := emailTask{
		ID:         taskID,
		TaskType:   TaskTypeTemplate,
		TmplEmail:  email,
		ResultChan: resultChan,
	}

	// Queue the task
	c.workerQueue <- task
	return taskID, resultChan
}

// SendTemplateByName sends an email using a named template from the configuration
func (c *Client) SendTemplateByName(name string, to []string, subject string, templateData map[string]interface{}) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap[name]
	if !ok {
		return uuid.Nil, nil, errors.New("template not found in configuration")
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// sendHTMLEmailSync sends an HTML email synchronously
func (c *Client) sendHTMLEmailSync(email *HTMLEmail) error {
	message := mail.NewV3Mail()

	from := mail.NewEmail(email.FromName, email.FromEmail)
	message.SetFrom(from)

	message.Subject = email.Subject

	if email.ReplyTo != "" {
		message.SetReplyTo(mail.NewEmail(email.ReplyName, email.ReplyTo))
	}

	// Add content
	p := mail.NewContent("text/plain", email.PlainText)
	h := mail.NewContent("text/html", email.HTMLContent)
	message.AddContent(p, h)

	// Add recipients
	personalization := mail.NewPersonalization()
	for _, recipient := range email.To {
		personalization.AddTos(mail.NewEmail("", recipient))
	}
	message.AddPersonalizations(personalization)

	// Send the email
	response, err := c.client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return errors.New(response.Body)
	}

	return nil
}

// sendTemplateEmailSync sends a template email synchronously
func (c *Client) sendTemplateEmailSync(email *TemplateEmail) error {
	message := mail.NewV3Mail()

	from := mail.NewEmail(email.FromName, email.FromEmail)
	message.SetFrom(from)
	message.Subject = email.Subject

	message.SetTemplateID(email.TemplateID)
	if email.ReplyTo != "" {
		message.SetReplyTo(mail.NewEmail(email.ReplyName, email.ReplyTo))
	}

	// Add recipients and template data
	personalization := mail.NewPersonalization()
	for _, recipient := range email.To {
		personalization.AddTos(mail.NewEmail("", recipient))
	}

	// Add dynamic template data
	if email.TemplateData != nil {
		personalization.DynamicTemplateData = email.TemplateData
	}

	message.AddPersonalizations(personalization)

	// Add attachments if any
	if len(email.Attachments) > 0 {
		for _, att := range email.Attachments {
			attachment := mail.NewAttachment()
			// SendGrid requires base64 encoded content
			encodedContent := base64.StdEncoding.EncodeToString(att.Content)
			attachment.SetContent(encodedContent)
			attachment.SetType(att.Type)
			attachment.SetFilename(att.Filename)
			if att.Disposition != "" {
				attachment.SetDisposition(att.Disposition)
			} else {
				attachment.SetDisposition("attachment")
			}
			if att.ContentID != "" {
				attachment.SetContentID(att.ContentID)
			}
			message.AddAttachment(attachment)
		}
	}

	// Send the email
	response, err := c.client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return errors.New(response.Body)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email using the welcome template
func (c *Client) SendWelcomeEmail(to []string, subject string, data WelcomeTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["welcome"]
	if !ok {
		return uuid.Nil, nil, errors.New("welcome template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"user_name":        data.UserName,
		"home_url":         data.HomeURL,
		"business_name":    data.BusinessName,
		"customer_logo":    data.CustomerLogo,
		"terms_conditions": data.TermsConditions,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendPasswordResetEmail sends a password reset email using the password reset template
func (c *Client) SendPasswordResetEmail(to []string, subject string, data PasswordResetTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["password_reset"]
	if !ok {
		return uuid.Nil, nil, errors.New("password reset template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"user_name":        data.UserName,
		"reset_url":        data.ResetURL,
		"terms_conditions": data.TermsConditions,
		"logo":             data.Logo,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendPaymentLinkEmail sends a payment link email using the payment_link template
func (c *Client) SendPaymentLinkEmail(to []string, subject string, data PaymentLinkTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["payment_link"]
	if !ok {
		return uuid.Nil, nil, errors.New("payment link template not found in configuration")
	}

	// Convert strongly typed data to a map matching SendGrid dynamic keys
	templateData := map[string]interface{}{
		"recipient":        data.Recipient,
		"sender":           data.Sender,
		"terms_conditions": data.TermsConditions,
		"name":             data.Name,
		"reply_name":       data.ReplyName,
		"custom_message":   data.CustomMessage,
		"payment_url":      data.PaymentURL,
		"invoice_id":       data.InvoiceID,
		"amount":           data.Amount,
		"logo":             data.Logo,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendMessageEmail sends a message email using the message template
func (c *Client) SendMessageEmail(to []string, subject string, data MessageTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["message"]
	if !ok {
		return uuid.Nil, nil, errors.New("message template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"content":          data.Content,
		"recipient":        data.Recipient,
		"sender":           data.Sender,
		"home_url":         data.HomeURL,
		"terms_conditions": data.TermsConditions,
		"logo":             data.Logo,
		"subject":          subject,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendMessageExternalEmail sends a message email using the message template
func (c *Client) SendExternalMessageEmail(to []string, subject string, data ExternalMessageTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["message_external"]
	if !ok {
		return uuid.Nil, nil, errors.New("message template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"content":          data.Content,
		"recipient":        data.Recipient,
		"sender":           data.Sender,
		"reply_to":         data.ReplyTo,
		"terms_conditions": data.TermsConditions,
		"logo":             data.Logo,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		ReplyTo:      data.ReplyTo,
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendInviteEmail sends an invite email using the invite template
func (c *Client) SendInviteEmail(to []string, subject string, data InviteTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["invite"]
	if !ok {
		return uuid.Nil, nil, errors.New("invite template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"user_name":        data.UserName,
		"invite_url":       data.InviteURL,
		"terms_conditions": data.TermsConditions,
		"logo":             data.Logo,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendInviteCustomerEmail sends an invite email using the invite template
func (c *Client) SendInviteCustomerEmail(to []string, subject string, data InviteCustomerTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["invite_customer"]
	if !ok {
		return uuid.Nil, nil, errors.New("invite template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"user_name":        data.UserName,
		"invite_url":       data.InviteURL,
		"terms_conditions": data.TermsConditions,
		"logo":             data.Logo,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendAssignedToMarinaEmail sends an email using the assigned_to_marina template
func (c *Client) SendAssignedToMarinaEmail(to []string, subject string, data AssignedToMarinaTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["assigned_to_marina"]
	if !ok {
		return uuid.Nil, nil, errors.New("assigned_to_marina template not found in configuration")
	}

	email := &TemplateEmail{
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID: templateID,
		TemplateData: map[string]interface{}{
			"customer_logo":    data.CustomerLogo,
			"business_name":    data.BusinessName,
			"user_name":        data.UserName,
			"home_url":         data.HomeURL,
			"terms_conditions": data.TermsConditions,
		},
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

// SendESignSubmissionEmail sends a message email using the message template
func (c *Client) SendESignSubmissionEmail(to []string, subject string, data ESignSubmissionTemplateData, attachments ...Attachment) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["esign_submission"]
	if !ok {
		return uuid.Nil, nil, errors.New("esign submission template not found in configuration")
	}

	// Convert the strongly typed data to a map
	templateData := map[string]interface{}{
		"recipient":        data.Recipient,
		"sender":           data.Sender,
		"document_url":     data.DocumentURL,
		"terms_conditions": data.TermsConditions,
		"reply_name":       data.ReplyName,
		"name":             data.Name,
		"custom_message":   data.CustomMessage,
		"logo":             data.Logo,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		ReplyTo:      data.ReplyTo,
		ReplyName:    data.ReplyName,
		TemplateID:   templateID,
		TemplateData: templateData,
		Attachments:  attachments,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}

func (c *Client) SendNotificationEmail(to []string, subject string, data NotificationTemplateData) (uuid.UUID, <-chan EmailStatus, error) {
	templateID, ok := c.config.TemplatesMap["notification"]
	if !ok {
		return uuid.Nil, nil, errors.New("notification template not found in configuration")
	}

	templateData := map[string]interface{}{
		"recipient":     data.Recipient,
		"type":          data.Type,
		"customer_name": data.CustomerName,
		"home_url":      data.HomeURL,
		"logo":          data.Logo,
		"subject":       subject,
	}

	email := &TemplateEmail{
		Subject: subject,
		EmailData: EmailData{
			To:        to,
			Subject:   subject,
			FromEmail: c.config.FromEmail,
			FromName:  c.config.FromName,
		},
		TemplateID:   templateID,
		TemplateData: templateData,
	}

	taskID, resultChan := c.SendTemplateEmail(email)
	return taskID, resultChan, nil
}
