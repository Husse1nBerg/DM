package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	tokenpkg "github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type InviteHandler struct {
	server *s.Server
}

func NewInviteHandler(server *s.Server) *InviteHandler {
	return &InviteHandler{server: server}
}

// ConfirmToken godoc
// @Summary Confirm an invitation token
// @Description Checks if the provided token is valid
// @Tags Invitations
// @Accept json
// @Produce json
// @Param token query string true "Invitation token"
// @Param email query string true "User email"
// @Success 200 {object} responses.ConfirmTokenResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 404 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /api/v1/invite/confirm [get]
func (h *InviteHandler) ConfirmToken(c echo.Context) error {
	logger := h.server.Logger
	queries := h.server.DB.Queries()
	ctx := c.Request().Context()

	// Get token and email from query parameters
	token := c.QueryParam("token")
	email := utils.LowerCase(c.QueryParam("email"))

	// Validate parameters
	if token == "" {
		logger.Zap.Error("Missing token parameter")
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "Token is required",
		})
	}

	if email == "" {
		logger.Zap.Error("Missing email parameter")
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "Email is required",
		})
	}

	// Get the invite from the database
	invite, err := queries.GetInviteByToken(ctx, token)
	if err != nil {
		logger.Zap.Error("Failed to get invite by token", err)
		return c.JSON(http.StatusNotFound, responses.BaseResponse{
			Message: "Invalid or expired invitation token",
		})
	}

	// Check if the token has expired (explicit check even though the DB query handles this)
	if invite.ExpiresAt.Time.Before(time.Now()) {
		logger.Zap.Error("Token has expired",
			"token", token,
			"expires_at", invite.ExpiresAt.Time)
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "Invitation token has expired",
		})
	}
	isUsed := invite.Used
	// check if invite is already used
	if *isUsed {
		logger.Zap.Error("Invitation token already used",
			"token", token)
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "Invitation token already used",
		})
	}
	// Check if the email matches
	if invite.Email != email {
		logger.Zap.Error("Email mismatch for invitation token",
			"token", token,
			"request_email", email,
			"invite_email", invite.Email)
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "Email does not match invitation",
		})
	}

	// Convert pgtype.Timestamp to time.Time for the response
	expiresAt := invite.ExpiresAt.Time

	return c.JSON(http.StatusOK, responses.ConfirmTokenResponse{
		BaseResponse: responses.BaseResponse{
			Message: "Token is valid",
		},
		Data: responses.InviteResponse{
			ID:        invite.ID.String(),
			Email:     invite.Email,
			ExpiresAt: expiresAt,
			Valid:     true,
		},
	})
}

// AcceptInvitation godoc
// @Summary Accept an invitation
// @Description Accepts an invitation and sets the user's password
// @Tags Invitations
// @Accept json
// @Produce json
// @Param request body requests.AcceptInvitationRequest true "Invitation acceptance request"
// @Success 200 {object} responses.AcceptInvitationResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 404 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /api/v1/invite/accept [post]
func (h *InviteHandler) AcceptInvitation(c echo.Context) error {
	logger := h.server.Logger
	ctx := c.Request().Context()

	// Parse and validate the request body
	req := new(requests.AcceptInvitationRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get queries
	queries := h.server.DB.Queries()

	// Get the invite first to check if it's valid
	invite, err := queries.GetInviteByToken(ctx, req.Token)
	if err != nil {
		logger.Zap.Error("Failed to get invite by token", err)
		return c.JSON(http.StatusNotFound, responses.BaseResponse{
			Message: "Invalid or expired invitation token",
		})
	}

	// Hash the password using the utility function from utils
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Zap.Error("Failed to hash password", err)
		return c.JSON(http.StatusInternalServerError, responses.BaseResponse{
			Message: "Failed to process password",
		})
	}

	// Update the user with the new password
	params := db.UpdateUserInviteParams{
		ID:           invite.UserID,
		PasswordHash: utils.Pointer(passwordHash),
	}

	_, err = queries.UpdateUserInvite(ctx, params)
	if err != nil {
		logger.Zap.Error("Failed to update user password", err)
		return c.JSON(http.StatusInternalServerError, responses.BaseResponse{
			Message: "Failed to update user information",
		})
	}
	// Mark the invitation as used
	_, err = queries.MarkInviteAsUsed(ctx, req.Token)
	if err != nil {
		logger.Zap.Error("Failed to mark invite as used", err)
		return c.JSON(http.StatusInternalServerError, responses.BaseResponse{
			Message: "Failed to process invitation",
		})
	}

	// Get the user's role to check if they are a customer (more efficient than full user query)
	userRole, err := queries.GetRoleByID(ctx, invite.UserID)
	if err != nil {
		logger.Zap.Error("Failed to get user role after invitation acceptance", err)
		// Don't fail the request, just log the error and continue
	} else {
		// Only send notifications if the user who accepted the invitation has a customer role
		if userRole.IsCustomerRole != nil && *userRole.IsCustomerRole {
			// Send notifications to marina staff asynchronously
			go func() {
				// Use background context with timeout for notification operations
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()

				// Get the user's basic info needed for notifications (minimal query)
				user, err := queries.GetUserByID(ctx, invite.UserID)
				if err != nil {
					logger.Zap.Warnw("Failed to get user info for notification", "user_id", invite.UserID, "error", err)
					return
				}

				// Get marina staff users (non-customers) for notifications
				marinaUsers, err := queries.GetUsersByMarina(ctx, db.GetUsersByMarinaParams{
					MarinaID:   user.MarinaID,
					IsCustomer: utils.Pointer(false), // Get marina staff, not customers
				})
				if err != nil {
					logger.Zap.Warnw("Failed to get marina users for invitation notification",
						"marina_id", user.MarinaID, "error", err)
					return
				}

				// Create notifications for marina staff using smart notification system
				emailData := &notifications.EmailNotificationData{
					To:      []string{}, // No specific email recipients for invitation notifications
					Subject: "New Customer User Joined",
				}

				// Create notification service instance
				notificationService := notifications.NewNotificationService(
					queries,
					h.server.Redis,
					h.server.Logger,
					h.server.SendGrid,
					h.server.Config,
				)

				// Get customer ID safely
				customerID := ""
				if user.CustomerID != nil {
					customerID = *user.CustomerID
				}

				results, err := notificationService.CreateBulkDocumentNotifications(
					ctx,
					marinaUsers,
					user.OrganizationID,
					user.MarinaID,
					"Customer Invitation Accepted", // Use a descriptive filename
					customerID,                     // Customer ID
					emailData,
				)
				if err != nil {
					logger.Zap.Warnw("Failed to create bulk invitation notifications", "error", err)
				} else {
					// Log notification results
					for _, result := range results {
						if len(result.Errors) > 0 {
							logger.Zap.Warnw("Invitation notification delivery had errors",
								"user_id", result.UserID,
								"errors", result.Errors)
						} else {
							logger.Zap.Infow("Invitation notification delivered successfully",
								"user_id", result.UserID,
								"system", result.SystemDelivered,
								"email", result.EmailDelivered)
						}
					}
				}
			}()
		}
	}

	return c.JSON(http.StatusOK, responses.AcceptInvitationResponse{
		BaseResponse: responses.BaseResponse{
			Message: "Invitation accepted successfully",
		},
	})
}

// RefreshInvite godoc
// @Summary Refresh an invitation
// @Description Expires existing invitations and creates a new one for the specified user
// @Tags Invitations
// @Accept json
// @Produce json
// @Param request body requests.RefreshInviteRequest true "Refresh invitation request"
// @Success 200 {object} responses.RefreshInviteResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 404 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Security ApiKeyAuth
// @Router /api/v1/invite/refresh [post]
func (h *InviteHandler) RefreshInvite(c echo.Context) error {
	logger := h.server.Logger
	cfg := h.server.Config
	ctx := c.Request().Context()

	// Parse and validate the request body
	req := new(requests.RefreshInviteRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get queries
	queries := h.server.DB.Queries()

	var user db.User
	var err error

	// Get user by ID or email
	if req.UserID != nil {
		user, err = queries.GetUserByID(ctx, *req.UserID)
		if err != nil {
			logger.Zap.Error("Failed to get user by ID", err)
			return c.JSON(http.StatusNotFound, responses.BaseResponse{
				Message: "User not found",
			})
		}
	} else if req.Email != nil {
		email := utils.LowerCase(*req.Email)
		user, err = queries.GetUserByEmail(ctx, email)
		if err != nil {
			logger.Zap.Error("Failed to get user by email", err)
			return c.JSON(http.StatusNotFound, responses.BaseResponse{
				Message: "User not found",
			})
		}
	} else {
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "Either userId or email must be provided",
		})
	}

	// Check if user has a password set (if they do, they don't need an invite)
	if user.PasswordHash != nil {
		return c.JSON(http.StatusBadRequest, responses.BaseResponse{
			Message: "User already has a password set",
		})
	}

	// Expire existing invites for this user
	err = queries.ExpireInvitesByUserID(ctx, user.ID)
	if err != nil {
		logger.Zap.Error("Failed to expire existing invites", err)
		return c.JSON(http.StatusInternalServerError, responses.BaseResponse{
			Message: "Failed to expire existing invites",
		})
	}

	// Generate new invitation token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		logger.Zap.Error("Failed to generate token", err)
		return c.JSON(http.StatusInternalServerError, responses.BaseResponse{
			Message: "Failed to generate invitation token",
		})
	}

	// Create new invitation record
	invite, err := queries.CreateInvite(ctx, db.CreateInviteParams{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: utils.PgTimeNowAdd(240 * time.Hour), // 10 days
	})
	if err != nil {
		logger.Zap.Error("Failed to create new invite", err)
		return c.JSON(http.StatusInternalServerError, responses.BaseResponse{
			Message: "Failed to create new invitation",
		})
	}

	// Get user info from JWT token instead of database query to determine customer status
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*tokenpkg.JwtCustomClaims)

	// Check if user is a customer to determine email type
	var inviteURL string
	var templateData interface{}
	var subject string

	if claims.IsCustomer != nil && *claims.IsCustomer {
		// Customer user - use customer invite template
		inviteURL = fmt.Sprintf("%s/%s?token=%s&email=%s",
			cfg.App.FrontendBaseURL,
			cfg.App.InvitationCustomerRoute,
			token,
			url.QueryEscape(user.Email))

		termsConditionsURL := fmt.Sprintf("%s/%s",
			cfg.App.FrontendBaseURL,
			cfg.App.TermsConditionsRoute,
		)

		templateData = sendgrid.InviteCustomerTemplateData{
			UserName:        user.FirstName,
			InviteURL:       inviteURL,
			TermsConditions: termsConditionsURL,
		}
		subject = "DockMaster Customer Portal Invite"
	} else {
		// Staff user - use regular invite template
		inviteURL = fmt.Sprintf("%s/%s?token=%s&email=%s",
			cfg.App.FrontendBaseURL,
			cfg.App.InvitationRoute,
			token,
			url.QueryEscape(user.Email))

		termsConditionsURL := fmt.Sprintf("%s/%s",
			cfg.App.FrontendBaseURL,
			cfg.App.TermsConditionsRoute,
		)

		templateData = sendgrid.InviteTemplateData{
			UserName:        user.FirstName,
			InviteURL:       inviteURL,
			TermsConditions: termsConditionsURL,
		}
		subject = "DockMaster Platform Invite"
	}

	// Send invitation email
	var taskID uuid.UUID
	var resultChan <-chan sendgrid.EmailStatus

	if claims.IsCustomer != nil && *claims.IsCustomer {
		taskID, resultChan, err = h.server.SendGrid.SendInviteCustomerEmail(
			[]string{user.Email},
			subject,
			templateData.(sendgrid.InviteCustomerTemplateData),
		)
	} else {
		taskID, resultChan, err = h.server.SendGrid.SendInviteEmail(
			[]string{user.Email},
			subject,
			templateData.(sendgrid.InviteTemplateData),
		)
	}

	if err != nil {
		logger.Zap.Errorw("Failed to send invite email", "error", err)
		// Don't return error - invite was created successfully, just email failed
	} else {
		logger.Zap.Infow("Invite email queued",
			"email", user.Email,
			"task_id", taskID.String())

		// Log the email attempt (non-blocking)
		go func() {
			result := <-resultChan
			if result.Status == sendgrid.StatusSent {
				logger.Zap.Infow("Invite email sent successfully",
					"email", user.Email,
					"task_id", result.ID.String())
			} else {
				logger.Zap.Errorw("Failed to send invite email",
					"email", user.Email,
					"task_id", result.ID.String(),
					"error", result.Error)
			}
		}()
	}

	// Return the new invite information
	return c.JSON(http.StatusOK, responses.RefreshInviteResponse{
		BaseResponse: responses.BaseResponse{
			Message: "Invitation refreshed successfully",
		},
		Data: responses.InviteResponse{
			ID:        invite.ID.String(),
			Email:     invite.Email,
			ExpiresAt: invite.ExpiresAt.Time,
			Valid:     true,
		},
	})
}
