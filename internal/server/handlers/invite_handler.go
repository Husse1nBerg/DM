package handlers

import (
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/utils"
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
	email := c.QueryParam("email")

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
	return c.JSON(http.StatusOK, responses.AcceptInvitationResponse{
		BaseResponse: responses.BaseResponse{
			Message: "Invitation accepted successfully",
		},
	})
}
