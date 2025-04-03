package responses

// LoginResponse represents authentication tokens returned on login and refresh
// @Description Authentication token response containing access token, refresh token and expiration
type LoginResponse struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Exp          int64  `json:"exp" example:"1714637422"`
}

// LoginResponseWrapper is purely for Swagger documentation
type LoginResponseWrapper struct {
	Data    LoginResponse `json:"data"`
	Message string        `json:"message,omitempty" example:"Login successful"`
}

// RegisterResponseWrapper is purely for Swagger documentation
type RegisterResponseWrapper struct {
	Data    UserResponse `json:"data"`
	Message string       `json:"message" example:"User created successfully"`
}

// NewLoginSuccessResponse creates a success response with login information
func NewLoginSuccessResponse(token, refreshToken string, exp int64) BaseResponse {
	return NewSuccessResponse(&LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Exp:          exp,
	})
}
