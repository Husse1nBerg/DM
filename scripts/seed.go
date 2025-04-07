package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"

	"golang.org/x/crypto/bcrypt"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/models"
	u "github.com/dockworks/dm-web-backend/pkg/utils"
)

func main() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(cfg)
	q := db.Queries()

	// Seed Organization
	org, err := q.GetOrganizationByEmail(ctx, "dmweb@dockmaster.com")
	if err != nil {
		if err == pgx.ErrNoRows {

			orgAddress, err := q.CreateAddress(ctx, sqlc.CreateAddressParams{})
			if err != nil {
				log.Fatalf("failed to create org address: %v", err)
			}
			org, err = q.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
				Email:     "dmweb@dockmaster.com",
				Name:      "Dockmaster Web Org",
				AddressID: orgAddress.ID,
				IsActive:  u.Pointer(true),
			})
			if err != nil {
				log.Fatalf("failed to create organization: %v", err)
			}
			log.Println("Created organization: Acme Corp")
		} else {
			log.Fatalf("failed to get organization: %v", err)
		}
	}

	// Seed Marina
	marina, err := q.GetMarinaByEmail(ctx, "marina@dockmaster.com")
	if err != nil {
		if err == pgx.ErrNoRows {

			marinaAddress, err := q.CreateAddress(ctx, sqlc.CreateAddressParams{})
			if err != nil {
				log.Fatalf("failed to create marina address: %v", err)
			}

			// Create working hours using the new model
			workingHours := models.DefaultWorkingHours()
			workingHoursBytes, err := workingHours.ToBytes()
			if err != nil {
				log.Fatalf("failed to create working hours: %v", err)
			}

			marina, err = q.CreateMarina(ctx, sqlc.CreateMarinaParams{
				Name:           "Dockmaster Web",
				Email:          "marina@dockmaster.com",
				IsActive:       u.Pointer(true),
				AddressID:      marinaAddress.ID,
				OrganizationID: org.ID,
				WorkingHours:   workingHoursBytes,
			})
			if err != nil {
				log.Fatalf("failed to create marina: %v", err)
			}
			log.Println("Created marina: Acme Marina")
		} else {
			log.Fatalf("failed to get marina: %v", err)
		}
	}

	// Seed Roles
	roles := []sqlc.CreateRoleParams{
		{
			Name:        "superuser",
			Description: u.Pointer("Platform Superuser with full system access"),
			Permissions: []byte(`{
			"all": true,
			"admin_panel": true,
			"manage_org": true,
			"manage_marina": true,
			"manage_users": true,
			"manage_roles": true
		}`),
			IsActive: u.Pointer(true),
		},
		{
			Name:        "org_admin",
			Description: u.Pointer("Organization Administrator"),
			Permissions: []byte(`{
			"manage_org": true,
			"manage_marina": true,
			"manage_users": true,
			"read": true,
			"create": true,
			"update": true,
			"delete": true
		}`),
			IsActive: u.Pointer(true),
		},
		{
			Name:        "marina_admin",
			Description: u.Pointer("Administrator at the marina level"),
			Permissions: []byte(`{
			"manage_marina": true,
			"manage_users": true,
			"read": true,
			"create": true,
			"update": true,
			"delete": true
		}`),
			IsActive: u.Pointer(true),
		},
		{
			Name:        "staff",
			Description: u.Pointer("Regular staff member with limited write access"),
			Permissions: []byte(`{
			"read": true,
			"create": true,
			"update": true
		}`),
			IsActive: u.Pointer(true),
		},
		{
			Name:        "viewer",
			Description: u.Pointer("Read-only access"),
			Permissions: []byte(`{
			"read": true
		}`),
			IsActive: u.Pointer(true),
		},
	}

	for _, r := range roles {
		_, err := q.GetRoleByName(ctx, r.Name)
		if err != nil {

			if err == pgx.ErrNoRows {
				_, err := q.CreateRole(ctx, r)
				if err != nil {
					log.Fatalf("failed to create role %s: %v", r.Name, err)
				}
				log.Printf("Created role: %s", r.Name)
			} else {
				log.Fatalf("failed to get role: %v", err)
			}
		} else {
			log.Printf("Role already exists:  %s", r.Name)
		}
	}

	// Get admin role for user creation
	role, err := q.GetRoleByName(ctx, "superuser")
	if err != nil {

		log.Fatalf("admin role not found after seed: %v", err)
	}

	// Seed Admin User
	user, err := q.GetUserByEmail(ctx, cfg.App.AdminEmail)
	if err != nil {
		if err == pgx.ErrNoRows {

			encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.App.AdminPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("failed to encrypt password: %v", err)
			}
			user, err = q.CreateUser(ctx, sqlc.CreateUserParams{
				Username:       "admin",
				FirstName:      "Andrew",
				LastName:       "Sameh",
				Email:          cfg.App.AdminEmail,
				PasswordHash:   string(encryptedPassword),
				OrganizationID: org.ID,
				RoleID:         role.ID,
				MarinaID:       marina.ID,
				IsSuperuser:    u.Pointer(true),
				IsActive:       u.Pointer(true),
			})
			if err != nil {
				log.Fatalf("failed to create admin user: %v", err)
			}
			log.Println("Created admin user: ", cfg.App.AdminEmail)
		} else {
			log.Fatalf("failed to get user: %v", err)
		}
	} else {
		log.Printf("User already exists:  %s", user.Email)
	}
	err = q.AssignUserToMarina(ctx, sqlc.AssignUserToMarinaParams{
		UserID:   user.ID,
		MarinaID: marina.ID,
	})
	if err != nil {
		log.Fatalf("failed to assign user to marina: %v", err)
	}
	log.Println("Seed completed successfully")
}
