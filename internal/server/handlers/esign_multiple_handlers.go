package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// sendEmailsToSignersSequentially sends emails to signers in order, only sending to the first signer initially
func (h *EsignHandler) sendEmailsToSignersSequentially(submission db.EsignSubmission, signers []db.EsignSubmissionSigner, marina db.Marina, req requests.CreateMultipleEsignSubmissionRequest) error {
	if len(signers) == 0 {
		return fmt.Errorf("no signers to send emails to")
	}

	// Only send email to the first signer initially
	firstSigner := signers[0]

	// Create email data for the first signer
	var replyTo string
	var replyName string
	if req.ReplyTo != nil && *req.ReplyTo != "" {
		replyTo = *req.ReplyTo
	} else {
		replyTo = marina.Email
	}

	var customMessage string
	if req.CustomMessage != nil && *req.CustomMessage != "" {
		customMessage = *req.CustomMessage
	} else {
		customMessage = ""
	}

	if req.ReplyName != nil && *req.ReplyName != "" {
		replyName = *req.ReplyName
	}

	// Include marina logo if available
	var logo string
	if marina.Image != nil && *marina.Image != "" {
		fullURL := utils.GetFullImageURL(marina.Image)
		if fullURL != nil {
			logo = *fullURL
		}
	}

	// Safely handle optional Name
	var submissionName string
	if submission.Name != nil {
		submissionName = *submission.Name
	} else {
		submissionName = ""
	}

	email := sendgrid.ESignSubmissionTemplateData{
		DocumentURL:     h.server.Config.App.EsignDocumentURL(submission.ID.String()),
		Recipient:       "",
		Sender:          marina.Name,
		ReplyTo:         replyTo,
		ReplyName:       replyName,
		TermsConditions: h.server.Config.App.TermsConditionsURL(),
		Name:            submissionName,
		CustomMessage:   customMessage,
		Logo:            logo,
	}
	to := []string{firstSigner.Email}
	subject := "New e-signature submission"

	// Send email asynchronously
	taskID, resultChan, err := h.server.SendGrid.SendESignSubmissionEmail(to, subject, email)
	if err != nil {
		return fmt.Errorf("failed to send email to first signer: %w", err)
	}

	// Log the task
	h.server.Logger.Zap.Infow("Email queued for first signer", "task_id", taskID.String(), "to", firstSigner.Email)

	// Process the result asynchronously to log success/failure
	go func() {
		result := <-resultChan
		if result.Status == sendgrid.StatusSent {
			h.server.Logger.Zap.Infow("Email sent successfully to first signer",
				"to", firstSigner.Email,
				"task_id", result.ID.String(),
				"message_id", result.ID,
				"status", result.Status)
		} else {
			h.server.Logger.Zap.Errorw("Failed to send email to first signer",
				"to", firstSigner.Email,
				"task_id", result.ID.String(),
				"error", result.Error)
		}
	}()

	return nil
}

// processNextSigner processes the next signer in the sequence when a signer completes signing
func (h *EsignHandler) processNextSigner(ctx context.Context, submissionID uuid.UUID) {
	// Get the next signer in order
	nextSigner, err := h.server.DB.Queries().GetNextSignerForSubmission(ctx, submissionID)
	if err != nil {
		h.server.Logger.Zap.Errorw("Error getting next signer for submission",
			"submission_id", submissionID,
			"error", err)
		return
	}

	// Get submission and marina info
	submission, err := h.server.DB.Queries().GetEsignSubmissionByID(ctx, submissionID)
	if err != nil {
		h.server.Logger.Zap.Errorw("Error getting submission for next signer",
			"submission_id", submissionID,
			"error", err)
		return
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, submission.MarinaID)
	if err != nil {
		h.server.Logger.Zap.Errorw("Error getting marina for next signer",
			"submission_id", submissionID,
			"error", err)
		return
	}

	// Create email data for the next signer
	var replyTo string
	if submission.ReplyTo != nil && *submission.ReplyTo != "" {
		replyTo = *submission.ReplyTo
	} else {
		replyTo = marina.Email
	}

	var customMessage string
	if submission.CustomMessage != nil && *submission.CustomMessage != "" {
		customMessage = *submission.CustomMessage
	} else {
		customMessage = ""
	}

	// Include marina logo if available
	var logo string
	if marina.Image != nil && *marina.Image != "" {
		fullURL := utils.GetFullImageURL(marina.Image)
		if fullURL != nil {
			logo = *fullURL
		}
	}

	// Safely handle optional Name
	var submissionName string
	if submission.Name != nil {
		submissionName = *submission.Name
	} else {
		submissionName = ""
	}

	email := sendgrid.ESignSubmissionTemplateData{
		DocumentURL:     h.server.Config.App.EsignDocumentURL(submission.ID.String()),
		Recipient:       "",
		Sender:          marina.Name,
		ReplyTo:         replyTo,
		ReplyName:       "", // Could be enhanced to store reply name
		TermsConditions: h.server.Config.App.TermsConditionsURL(),
		Name:            submissionName,
		CustomMessage:   customMessage,
		Logo:            logo,
	}
	to := []string{nextSigner.Email}
	subject := "E-signature document ready for your signature"

	// Send email asynchronously
	taskID, resultChan, err := h.server.SendGrid.SendESignSubmissionEmail(to, subject, email)
	if err != nil {
		h.server.Logger.Zap.Errorw("Failed to send email to next signer",
			"submission_id", submissionID,
			"signer_email", nextSigner.Email,
			"error", err)
		return
	}

	// Log the task
	h.server.Logger.Zap.Infow("Email queued for next signer",
		"task_id", taskID.String(),
		"to", nextSigner.Email,
		"submission_id", submissionID)

	// Process the result asynchronously to log success/failure
	go func() {
		result := <-resultChan
		if result.Status == sendgrid.StatusSent {
			h.server.Logger.Zap.Infow("Email sent successfully to next signer",
				"to", nextSigner.Email,
				"task_id", result.ID.String(),
				"message_id", result.ID,
				"status", result.Status,
				"submission_id", submissionID)
		} else {
			h.server.Logger.Zap.Errorw("Failed to send email to next signer",
				"to", nextSigner.Email,
				"task_id", result.ID.String(),
				"error", result.Error,
				"submission_id", submissionID)
		}
	}()
}

// CreateMultipleEsignSubmission creates a new multiple e-signature submission
//
//	@Summary		Create multiple e-signature submission
//	@Description	Creates a new e-signature submission with multiple signers
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateMultipleEsignSubmissionRequest	true	"Multiple e-signature submission data"
//	@Success		201		{object}	responses.BaseResponse{data=responses.EsignSubmissionWithSignersResponse}
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		401		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions/multiple [post]
func (h *EsignHandler) CreateMultipleEsignSubmission(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	// Parse request body
	var req requests.CreateMultipleEsignSubmissionRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Error parsing request body", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request body").JSON(c)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Error validating request", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed: "+err.Error()).JSON(c)
	}

	// Validate that we have at least 2 signers
	if len(req.Signers) < 2 {
		return responses.NewErrorResponse(http.StatusBadRequest, "At least 2 signers are required for multiple signature submissions").JSON(c)
	}

	// Validate sign order uniqueness
	signOrders := make(map[int32]bool)
	for _, signer := range req.Signers {
		if signOrders[signer.SignOrder] {
			return responses.NewErrorResponse(http.StatusBadRequest, "Sign order must be unique for each signer").JSON(c)
		}
		signOrders[signer.SignOrder] = true
	}

	// Get the document to duplicate
	document, err := h.server.DB.Queries().GetEsignDocumentByID(c.Request().Context(), req.DocumentID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching document by ID", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	// Duplicate the document file in S3
	duplicatedFilePath, err := h.esignService.DuplicateFile(c.Request().Context(), document.BlobUrl)
	if err != nil {
		h.server.Logger.Zap.Error("Error duplicating document file", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error duplicating document file: "+err.Error()).JSON(c)
	}

	// Create submission with duplicated file
	submission, err := h.server.DB.Queries().CreateEsignSubmission(c.Request().Context(), db.CreateEsignSubmissionParams{
		OrganizationID:      organizationID,
		MarinaID:            marinaID,
		DocumentID:          req.DocumentID,
		Status:              "pending", // Default status
		BlobUrl:             duplicatedFilePath,
		BlobMetadata:        nil, // Ignoring blob metadata for now as requested
		CustomerID:          req.CustomerID,
		Email:               req.Signers[0].Email, // Use first signer's email as primary
		Name:                req.Name,
		AttachmentRequired:  req.AttachmentRequired,
		ReplyTo:             req.ReplyTo,
		CustomMessage:       req.CustomMessage,
		IsMultipleSignature: true,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating e-signature submission", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating submission").JSON(c)
	}

	// Create signers
	var signers []db.EsignSubmissionSigner
	for _, signerReq := range req.Signers {
		signer, err := h.server.DB.Queries().CreateEsignSubmissionSigner(c.Request().Context(), db.CreateEsignSubmissionSignerParams{
			SubmissionID: submission.ID,
			Email:        signerReq.Email,
			Name:         signerReq.Name,
			SignOrder:    signerReq.SignOrder,
			Status:       "pending",
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error creating e-signature submission signer", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating signer").JSON(c)
		}
		signers = append(signers, signer)
	}

	// Increment document usage immediately after successful submission
	_, err = h.server.DB.Queries().IncrementMarinaDocumentUsage(c.Request().Context(), db.IncrementMarinaDocumentUsageParams{
		ID:      marinaID,
		Column2: 1,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error incrementing document usage", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error incrementing document usage").JSON(c)
	}

	// Fetch marina again for email sender info
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}

	// Send emails to signers sequentially
	err = h.sendEmailsToSignersSequentially(submission, signers, marina, req)
	if err != nil {
		h.server.Logger.Zap.Error("Error sending emails to signers", err)
		// Don't fail the request, just log the error
	}

	response := responses.NewEsignSubmissionWithSignersResponseSuccess(submission, signers)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetEsignSubmissionSigners retrieves all signers for a submission
//
//	@Summary		Get e-signature submission signers
//	@Description	Retrieves all signers for an e-signature submission
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Submission ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=[]responses.EsignSubmissionSignerResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Router			/public/esign/submissions/{id}/signers [get]
func (h *EsignHandler) GetEsignSubmissionSignersPublic(c echo.Context) error {

	// Parse submission ID
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Get signers
	signers, err := h.server.DB.Queries().ListEsignSubmissionSignersBySubmissionID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submission signers", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Signers not found").JSON(c)
	}

	return responses.NewEsignSubmissionSignersResponseSuccess(signers).JSON(c)
}

// UpdateEsignSubmissionSigner updates a signer's status
//
//	@Summary		Update e-signature submission signer
//	@Description	Updates a signer's status in an e-signature submission
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string											true	"Signer ID"	Format(uuid)
//	@Param			request	body		requests.UpdateEsignSubmissionSignerRequest	true	"Signer update data"
//	@Success		200		{object}	responses.BaseResponse{data=responses.EsignSubmissionSignerResponse}
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		401		{object}	responses.BaseResponse
//	@Failure		404		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Router			/public/esign/submissions/signers/{id} [put]
func (h *EsignHandler) UpdateEsignSubmissionSignerPublic(c echo.Context) error {

	// Parse signer ID
	signerIDStr := c.Param("id")
	signerID, err := uuid.Parse(signerIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid signer ID format").JSON(c)
	}

	// Parse request body
	var req requests.UpdateEsignSubmissionSignerRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Error parsing request body", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request body").JSON(c)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Error validating request", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed: "+err.Error()).JSON(c)
	}

	// Update signer status based on the status type
	var signer db.EsignSubmissionSigner

	switch req.Status {
	case "signed":
		signer, err = h.server.DB.Queries().UpdateEsignSubmissionSignerToSigned(c.Request().Context(), signerID)
	case "declined":
		signer, err = h.server.DB.Queries().UpdateEsignSubmissionSignerToDeclined(c.Request().Context(), db.UpdateEsignSubmissionSignerToDeclinedParams{
			ID:             signerID,
			DeclinedReason: req.DeclinedReason,
		})
	default:
		// For other statuses like "pending", "expired", etc.
		signer, err = h.server.DB.Queries().UpdateEsignSubmissionSignerStatus(c.Request().Context(), db.UpdateEsignSubmissionSignerStatusParams{
			ID:     signerID,
			Status: req.Status,
		})
	}

	if err != nil {
		h.server.Logger.Zap.Error("Error updating e-signature submission signer", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Signer not found").JSON(c)
	}

	// If signer signed, check if we need to send email to next signer
	if req.Status == "signed" {
		go h.processNextSigner(context.Background(), signer.SubmissionID)
	}

	return responses.NewEsignSubmissionSignerResponseSuccess(signer).JSON(c)
}

// GetEsignSubmissionWithSigners retrieves a submission with all its signers
//
//	@Summary		Get e-signature submission with signers
//	@Description	Retrieves an e-signature submission with all its signers
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Submission ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.EsignSubmissionWithSignersResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Router			/public/esign/submissions/{id}/with-signers [get]
func (h *EsignHandler) GetEsignSubmissionWithSignersPublic(c echo.Context) error {

	// Parse submission ID
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Get submission
	submission, err := h.server.DB.Queries().GetEsignSubmissionByID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submission", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Submission not found").JSON(c)
	}

	// Get signers
	signers, err := h.server.DB.Queries().ListEsignSubmissionSignersBySubmissionID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submission signers", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Signers not found").JSON(c)
	}

	return responses.NewEsignSubmissionWithSignersResponseSuccess(submission, signers).JSON(c)
}
