package token

import (
	"testing"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestCreateAccessToken_IncludesIsCustomer(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			AccessSecret: "test-secret-key-for-testing-only",
		},
	}

	// Create a test user with IsCustomer set to true
	user := &db.User{
		ID:             uuid.New(),
		Username:       "testuser",
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@example.com",
		OrganizationID: uuid.New(),
		MarinaID:       uuid.New(),
		RoleID:         uuid.New(),
		IsCustomer:     &[]bool{true}[0], // Pointer to true
	}

	// Create token service
	service := NewTokenService(cfg)

	// Create access token
	tokenString, _, err := service.CreateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	// Parse the token to verify claims
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Auth.AccessSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	// Verify IsCustomer field is included and correct
	if claims.IsCustomer == nil {
		t.Error("IsCustomer field is nil in token claims")
	} else if !*claims.IsCustomer {
		t.Error("IsCustomer field is false, expected true")
	}

	// Verify other fields are also correct
	if claims.ID != user.ID {
		t.Errorf("ID mismatch: got %v, want %v", claims.ID, user.ID)
	}
	if claims.Email != user.Email {
		t.Errorf("Email mismatch: got %v, want %v", claims.Email, user.Email)
	}
}

func TestCreateAccessToken_IsCustomerFalse(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			AccessSecret: "test-secret-key-for-testing-only",
		},
	}

	// Create a test user with IsCustomer set to false
	user := &db.User{
		ID:             uuid.New(),
		Username:       "testuser",
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@example.com",
		OrganizationID: uuid.New(),
		MarinaID:       uuid.New(),
		RoleID:         uuid.New(),
		IsCustomer:     &[]bool{false}[0], // Pointer to false
	}

	// Create token service
	service := NewTokenService(cfg)

	// Create access token
	tokenString, _, err := service.CreateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	// Parse the token to verify claims
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Auth.AccessSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	// Verify IsCustomer field is included and correct
	if claims.IsCustomer == nil {
		t.Error("IsCustomer field is nil in token claims")
	} else if *claims.IsCustomer {
		t.Error("IsCustomer field is true, expected false")
	}
}

func TestCreateAccessToken_IsCustomerNil(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Auth: config.AuthConfig{
			AccessSecret: "test-secret-key-for-testing-only",
		},
	}

	// Create a test user with IsCustomer set to nil
	user := &db.User{
		ID:             uuid.New(),
		Username:       "testuser",
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@example.com",
		OrganizationID: uuid.New(),
		MarinaID:       uuid.New(),
		RoleID:         uuid.New(),
		IsCustomer:     nil, // nil value
	}

	// Create token service
	service := NewTokenService(cfg)

	// Create access token
	tokenString, _, err := service.CreateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to create access token: %v", err)
	}

	// Parse the token to verify claims
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Auth.AccessSecret), nil
	})
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		t.Fatal("Failed to extract claims")
	}

	// Verify IsCustomer field is nil
	if claims.IsCustomer != nil {
		t.Error("IsCustomer field is not nil, expected nil")
	}
}
