package handlers

import (
	"fmt"
	"net/http"
	"time"

	database "github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	tokenservice "github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"

	"github.com/labstack/echo/v4"

	jwtGo "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthHandler struct {
	server *s.Server
}

func NewAuthHandler(server *s.Server) *AuthHandler {
	return &AuthHandler{server: server}
}

// Login
//
//	@Summary		Authenticate a user
//	@Description	Perform user login
//	@ID				user-login
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.LoginRequest	true	"User's credentials"
//	@Success		200		{object}	responses.LoginResponseWrapper	"Success response with login data"
//	@Failure		400		{object}	responses.Error					"Validation error"
//	@Failure		401		{object}	responses.Error					"Authentication error"
//	@Failure		403		{object}	responses.Error					"Account locked"
//	@Failure		500		{object}	responses.Error					"Server error"
//	@Router			/auth/login [post]
func (authHandler *AuthHandler) Login(c echo.Context) error {
	logger := authHandler.server.Logger
	queries := authHandler.server.DB.Queries()
	ctx := c.Request().Context()

	loginRequest := new(requests.LoginRequest)

	logger.LogWithFields("User is trying to login", c.Response().Header().Get(echo.HeaderXRequestID), "auth")

	if err := c.Bind(loginRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(loginRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	user, err := queries.GetUserByEmail(ctx, loginRequest.Email)

	if err != nil {
		logger.Zap.Info("login failed: user not found ", err, loginRequest.Email, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid credentials").JSON(c)
	}

	// Check if account is locked
	now := time.Now().UTC()

	if user.LockedUntil.Valid && user.LockedUntil.Time.After(now) {
		logger.Zap.Info("login failed: account locked", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusForbidden, "Account is locked due to too many failed login attempts. Please try again later.").JSON(c)
	}

	// Check password
	if err := utils.VerifyPassword(user.PasswordHash, loginRequest.Password); err != nil {
		logger.Zap.Info("login failed: invalid password", c.Response().Header().Get(echo.HeaderXRequestID))

		// Increment failed login attempts
		attempts := int32(1)
		if user.FailedLoginAttempts != nil {
			attempts = *user.FailedLoginAttempts + 1
		}

		// Initialize the lockedUntil variable as not locked by default
		var lockedUntil pgtype.Timestamp

		// Check if account should be locked
		maxAttempts := authHandler.server.Config.Auth.LoginAttempts
		lockoutMins := authHandler.server.Config.Auth.LockoutDuration

		if attempts >= maxAttempts {
			// Lock the account
			lockUntil := now.Add(time.Duration(lockoutMins) * time.Minute)
			// Create a timestamp using a Time value and setting Valid to true
			lockedUntil = pgtype.Timestamp{
				Time:  lockUntil,
				Valid: true,
			}
			logger.Zap.Info("account locked due to too many failed attempts", c.Response().Header().Get(echo.HeaderXRequestID))
		}

		// Update user record with incremented attempts and possible lock
		_, updateErr := queries.UpdateUser(ctx, database.UpdateUserParams{
			ID:                  user.ID,
			FirstName:           user.FirstName,
			LastName:            user.LastName,
			Email:               user.Email,
			EmailVerified:       user.EmailVerified,
			Phone:               user.Phone,
			Title:               user.Title,
			Image:               user.Image,
			PasswordHash:        user.PasswordHash,
			LastLogin:           user.LastLogin,
			FailedLoginAttempts: &attempts,
			LockedUntil:         lockedUntil,
			LastPasswordReset:   user.LastPasswordReset,
			MarinaID:            user.MarinaID,
			RoleID:              user.RoleID,
			IsSuperuser:         user.IsSuperuser,
			IsActive:            user.IsActive,
		})

		if updateErr != nil {
			logger.Zap.Error("failed to update login attempts", updateErr, c.Response().Header().Get(echo.HeaderXRequestID))
		}

		// If we just locked the account, return a 403 instead of 401
		if attempts >= maxAttempts {
			return responses.NewErrorResponse(http.StatusForbidden, "Account is locked due to too many failed login attempts. Please try again later.").JSON(c)
		}

		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid credentials").JSON(c)
	}

	// At this point, the user is not locked and the password is correct
	// Successful login - reset failed attempts counter
	resetAttempts := int32(0)

	// Update user's last login time and reset failed attempts, but preserve lock status
	_, updateErr := queries.UpdateUser(ctx, database.UpdateUserParams{
		ID:                  user.ID,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		EmailVerified:       user.EmailVerified,
		Phone:               user.Phone,
		Title:               user.Title,
		Image:               user.Image,
		PasswordHash:        user.PasswordHash,
		LastLogin:           pgtype.Timestamp{Time: time.Now(), Valid: true},
		FailedLoginAttempts: &resetAttempts,
		LockedUntil:         user.LockedUntil, // Preserve lock status instead of resetting it
		LastPasswordReset:   user.LastPasswordReset,
		MarinaID:            user.MarinaID,
		RoleID:              user.RoleID,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
	})

	if updateErr != nil {
		logger.Zap.Error("failed to update user login data", updateErr, c.Response().Header().Get(echo.HeaderXRequestID))
		// Continue processing despite error to not affect user experience
	}

	// Recheck account lock after update - in case it got locked in another concurrent session
	updatedUser, err := queries.GetUserByID(ctx, user.ID)
	if err != nil {
		logger.Zap.Error("failed to get updated user data", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error retrieving user data").JSON(c)
	}

	// Double-check lock status
	if updatedUser.LockedUntil.Valid && updatedUser.LockedUntil.Time.After(now) {
		logger.Zap.Info("login rejected: account is locked", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusForbidden, "Account is locked due to too many failed login attempts. Please try again later.").JSON(c)
	}

	tokenService := tokenservice.NewTokenService(authHandler.server.Config)
	accessToken, exp, err := tokenService.CreateAccessToken(&updatedUser)
	if err != nil {
		logger.Zap.Error("failed to create access token: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error generating authentication token").JSON(c)
	}
	refreshToken, err := tokenService.CreateRefreshToken(&updatedUser)
	if err != nil {
		logger.Zap.Error("failed to create refresh token: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error generating refresh token").JSON(c)
	}

	return responses.NewLoginSuccessResponse(accessToken, refreshToken, exp).JSON(c)
}

// RefreshToken
//
//	@Summary		Refresh access token
//	@Description	Perform refresh access token
//	@ID				user-refresh
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	responses.LoginResponseWrapper	"Success response with new tokens"
//	@Failure		400		{object}	responses.Error					"Validation error"
//	@Failure		401		{object}	responses.Error					"Authentication error"
//	@Failure		403		{object}	responses.Error					"Account locked"
//	@Failure		500		{object}	responses.Error					"Server error"
//	@Router			/auth/refresh [post]
func (authHandler *AuthHandler) RefreshToken(c echo.Context) error {
	logger := authHandler.server.Logger
	queries := authHandler.server.DB.Queries()

	refreshRequest := new(requests.RefreshRequest)

	if err := c.Bind(refreshRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(refreshRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	token, err := jwtGo.Parse(refreshRequest.Token, func(token *jwtGo.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtGo.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(authHandler.server.Config.Auth.RefreshSecret), nil
	})

	if err != nil {
		logger.Zap.Info("token refresh failed: invalid token", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	claims, ok := token.Claims.(jwtGo.MapClaims)
	if !ok && !token.Valid {
		logger.Zap.Info("token refresh failed: invalid token claims", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	userID, err := uuid.Parse(claims["id"].(string))
	if err != nil {
		logger.Zap.Error("failed to parse user ID from token: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	user, err := queries.GetUserByID(c.Request().Context(), userID)

	if err != nil {
		logger.Zap.Info("token refresh failed: user not found", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	// Check if account is locked
	now := time.Now()
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(now) {
		logger.Zap.Info("token refresh failed: account locked", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusForbidden, "Account is locked due to too many failed login attempts. Please try again later.").JSON(c)
	}

	tokenService := tokenservice.NewTokenService(authHandler.server.Config)
	accessToken, exp, err := tokenService.CreateAccessToken(&user)
	if err != nil {
		logger.Zap.Error("failed to create access token: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error generating authentication token").JSON(c)
	}
	refreshToken, err := tokenService.CreateRefreshToken(&user)
	if err != nil {
		logger.Zap.Error("failed to create refresh token: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error generating refresh token").JSON(c)
	}

	return responses.NewLoginSuccessResponse(accessToken, refreshToken, exp).JSON(c)
}

// Register
//
//	@Summary		Register
//	@Description	New user registration
//	@ID				user-register
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.RegisterRequest	true	"User's registration details"
//	@Success		201		{object}	responses.RegisterResponseWrapper	"User created successfully"
//	@Failure		400		{object}	responses.Error					"Validation error"
//	@Failure		500		{object}	responses.Error					"Server error"
//	@Router			/auth/register [post]
func (authHandler *AuthHandler) Register(c echo.Context) error {
	logger := authHandler.server.Logger
	queries := authHandler.server.DB.Queries()

	registerRequest := new(requests.RegisterRequest)

	// Binding (will also check for unknown fields)
	if err := c.Bind(registerRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Explicit validation after binding
	if err := c.Validate(registerRequest); err != nil {
		// The NewErrorResponse function will properly format validation errors
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Check if email already exists
	_, err := queries.GetUserByEmail(c.Request().Context(), registerRequest.Email)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Email already in use").JSON(c)
	}

	// Check if username already exists
	_, err = queries.GetUserByUsername(c.Request().Context(), registerRequest.Username)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Username already in use").JSON(c)
	}

	// Hash the password
	encryptedPassword, err := utils.HashPassword(registerRequest.Password)
	if err != nil {
		logger.Zap.Error("failed to encrypt password: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing registration").JSON(c)
	}

	// Get organization ID - using first org for now (in a real app, you'd handle this differently)
	orgs, err := queries.GetAllOrganizations(c.Request().Context())
	if err != nil || len(orgs) == 0 {
		logger.Zap.Error("failed to get organizations: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to assign organization").JSON(c)
	}
	organizationID := orgs[0].ID

	// Get marinas for the organization
	marinas, err := queries.GetMarinasByOrganization(c.Request().Context(), organizationID)
	if err != nil || len(marinas) == 0 {
		logger.Zap.Error("failed to get marinas: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to assign marina").JSON(c)
	}
	marinaID := marinas[0].ID

	// Default to active user
	isActive := true

	// Create user parameters with all required fields
	userParams := database.CreateUserParams{
		FirstName:      registerRequest.FirstName,
		LastName:       registerRequest.LastName,
		Username:       registerRequest.Username,
		Email:          registerRequest.Email,
		RoleID:         registerRequest.RoleID,
		PasswordHash:   string(encryptedPassword),
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		IsActive:       &isActive,
	}

	newUser, err := queries.CreateUser(c.Request().Context(), userParams)
	if err != nil {
		logger.Zap.Error("failed to create user: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create user").JSON(c)
	}

	// Assign user to marina
	err = queries.AssignUserToMarina(c.Request().Context(), database.AssignUserToMarinaParams{
		UserID:   newUser.ID,
		MarinaID: marinaID,
	})
	if err != nil {
		logger.Zap.Error("failed to assign user to marina: %v", err, c.Response().Header().Get(echo.HeaderXRequestID))
		// We continue anyway since the user was created successfully
	}

	logger.LogWithFields("User registered successfully", c.Response().Header().Get(echo.HeaderXRequestID), "auth")

	// Use our new response structure that hides the password
	successResponse := responses.NewUserResponseSuccess(newUser)
	successResponse.Code = http.StatusCreated
	successResponse.Message = "User created successfully"

	return successResponse.JSON(c)
}
