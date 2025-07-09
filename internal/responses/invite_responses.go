package responses

import "time"

type InviteResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	Valid     bool      `json:"valid"`
}

type ConfirmTokenResponse struct {
	BaseResponse
	Data InviteResponse `json:"data"`
}

type AcceptInvitationResponse struct {
	BaseResponse
}

type RefreshInviteResponse struct {
	BaseResponse
	Data InviteResponse `json:"data"`
}
