package token

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	db "github.com/dockworks/dm-web-backend/internal/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const ExpireCount = 2
const ExpireRefreshCount = 168

type JwtCustomClaims struct {
	ID       uuid.UUID `json:"id"`
	OrgId    uuid.UUID `json:"organizationId"`
	MarinaId uuid.UUID `json:"marinaId"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	RoleID   uuid.UUID `json:"roleId"`
	jwt.RegisteredClaims
}

type JwtCustomRefreshClaims struct {
	ID uuid.UUID `json:"id"`
	jwt.RegisteredClaims
}

type ServiceWrapper interface {
	CreateAccessToken(user *db.User) (accessToken string, exp int64, err error)
	CreateRefreshToken(user *db.User) (t string, err error)
}

type Service struct {
	config *config.Config
}

func NewTokenService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
	}
}

func (tokenService *Service) CreateAccessToken(user *db.User) (t string, expired int64, err error) {
	exp := time.Now().Add(time.Hour * ExpireCount)
	claims := &JwtCustomClaims{
		user.ID,
		user.OrganizationID,
		user.MarinaID,
		user.FirstName + " " + user.LastName,
		user.Email,
		user.RoleID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	expired = exp.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err = token.SignedString([]byte(tokenService.config.Auth.AccessSecret))
	if err != nil {
		return
	}

	return
}

func (tokenService *Service) CreateRefreshToken(user *db.User) (t string, err error) {
	claimsRefresh := &JwtCustomRefreshClaims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * ExpireRefreshCount)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)

	rt, err := refreshToken.SignedString([]byte(tokenService.config.Auth.RefreshSecret))
	if err != nil {
		return "", err
	}
	return rt, err
}
