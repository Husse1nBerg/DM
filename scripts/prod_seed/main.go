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

// RunProdSeed seeds the production/development database with initial data
func RunProdSeed() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.DB)
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

	// Seed Roles with new permission structure
	roles := []struct {
		name        string
		description string
		permissions *models.Permissions
		isActive    bool
	}{
		{
			name:        "superuser",
			description: "Platform Superuser with full system access",
			permissions: &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          true,
				DeleteUsers:         true,
				ReadOrganizations:   true,
				WriteOrganizations:  true,
				DeleteOrganizations: true,
				ReadMarinas:         true,
				WriteMarinas:        true,
				DeleteMarinas:       true,
				ReadRoles:           true,
				WriteRoles:          true,
				DeleteRoles:         true,
				ReadSettings:        true,
				WriteSettings:       true,
			},
			isActive: true,
		},
		{
			name:        "org_admin",
			description: "Organization Administrator",
			permissions: &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          true,
				DeleteUsers:         true,
				ReadOrganizations:   true,
				WriteOrganizations:  true,
				DeleteOrganizations: false,
				ReadMarinas:         true,
				WriteMarinas:        true,
				DeleteMarinas:       true,
				ReadRoles:           true,
				WriteRoles:          false,
				DeleteRoles:         false,
				ReadSettings:        true,
				WriteSettings:       true,
			},
			isActive: true,
		},
		{
			name:        "marina_admin",
			description: "Administrator at the marina level",
			permissions: &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          true,
				DeleteUsers:         false,
				ReadOrganizations:   true,
				WriteOrganizations:  false,
				DeleteOrganizations: false,
				ReadMarinas:         true,
				WriteMarinas:        true,
				DeleteMarinas:       false,
				ReadRoles:           true,
				WriteRoles:          false,
				DeleteRoles:         false,
				ReadSettings:        true,
				WriteSettings:       true,
			},
			isActive: true,
		},
		{
			name:        "staff",
			description: "Regular staff member with limited write access",
			permissions: &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          false,
				DeleteUsers:         false,
				ReadOrganizations:   true,
				WriteOrganizations:  false,
				DeleteOrganizations: false,
				ReadMarinas:         true,
				WriteMarinas:        true,
				DeleteMarinas:       false,
				ReadRoles:           true,
				WriteRoles:          false,
				DeleteRoles:         false,
				ReadSettings:        true,
				WriteSettings:       false,
			},
			isActive: true,
		},
		{
			name:        "viewer",
			description: "Read-only access",
			permissions: &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          false,
				DeleteUsers:         false,
				ReadOrganizations:   true,
				WriteOrganizations:  false,
				DeleteOrganizations: false,
				ReadMarinas:         true,
				WriteMarinas:        false,
				DeleteMarinas:       false,
				ReadRoles:           true,
				WriteRoles:          false,
				DeleteRoles:         false,
				ReadSettings:        true,
				WriteSettings:       false,
			},
			isActive: true,
		},
	}

	for _, r := range roles {
		// Convert permissions to bytes
		permBytes, err := r.permissions.ToBytes()
		if err != nil {
			log.Fatalf("failed to convert permissions to bytes for role %s: %v", r.name, err)
		}

		existingRole, err := q.GetRoleByName(ctx, r.name)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Create new role if doesn't exist
				_, err = q.CreateRole(ctx, sqlc.CreateRoleParams{
					Name:        r.name,
					Description: u.Pointer(r.description),
					Permissions: permBytes,
					IsActive:    u.Pointer(r.isActive),
				})
				if err != nil {
					log.Fatalf("failed to create role %s: %v", r.name, err)
				}
				log.Printf("Created role: %s", r.name)
			} else {
				log.Fatalf("failed to get role: %v", err)
			}
		} else {
			// Update existing role with new permissions
			_, err = q.UpdateRole(ctx, sqlc.UpdateRoleParams{
				ID:          existingRole.ID,
				Name:        r.name,
				Description: u.Pointer(r.description),
				Permissions: permBytes,
				IsActive:    u.Pointer(r.isActive),
			})
			if err != nil {
				log.Fatalf("failed to update role %s: %v", r.name, err)
			}
			log.Printf("Updated role: %s with new permission structure", r.name)
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

	// Seed Staff Role if it doesn't exist
	_, err = q.GetRoleByName(ctx, "staff")
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create staff role
			staffPermissions := &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          false,
				DeleteUsers:         false,
				ReadOrganizations:   true,
				WriteOrganizations:  false,
				DeleteOrganizations: false,
				ReadMarinas:         true,
				WriteMarinas:        true,
				DeleteMarinas:       false,
				ReadRoles:           true,
				WriteRoles:          false,
				DeleteRoles:         false,
				ReadSettings:        true,
				WriteSettings:       false,
			}
			permBytes, err := staffPermissions.ToBytes()
			if err != nil {
				log.Fatalf("failed to convert permissions to bytes for staff role: %v", err)
			}
			_, err = q.CreateRole(ctx, sqlc.CreateRoleParams{
				Name:        "staff",
				Description: u.Pointer("Regular staff member with limited write access"),
				Permissions: permBytes,
				IsActive:    u.Pointer(true),
			})
			if err != nil {
				log.Fatalf("failed to create staff role: %v", err)
			}
			log.Printf("Created staff role")
		} else {
			log.Fatalf("failed to get staff role: %v", err)
		}
	}

	// Seed Viewer Role if it doesn't exist
	_, err = q.GetRoleByName(ctx, "viewer")
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create viewer role
			viewerPermissions := &models.Permissions{
				ReadUsers:           true,
				WriteUsers:          false,
				DeleteUsers:         false,
				ReadOrganizations:   true,
				WriteOrganizations:  false,
				DeleteOrganizations: false,
				ReadMarinas:         true,
				WriteMarinas:        false,
				DeleteMarinas:       false,
				ReadRoles:           true,
				WriteRoles:          false,
				DeleteRoles:         false,
				ReadSettings:        true,
				WriteSettings:       false,
			}
			permBytes, err := viewerPermissions.ToBytes()
			if err != nil {
				log.Fatalf("failed to convert permissions to bytes for viewer role: %v", err)
			}
			_, err = q.CreateRole(ctx, sqlc.CreateRoleParams{
				Name:        "viewer",
				Description: u.Pointer("Read-only access"),
				Permissions: permBytes,
				IsActive:    u.Pointer(true),
			})
			if err != nil {
				log.Fatalf("failed to create viewer role: %v", err)
			}
			log.Printf("Created viewer role")
		} else {
			log.Fatalf("failed to get viewer role: %v", err)
		}
	}
}

func main() {
	log.Println("Starting production/development database seeding...")
	RunProdSeed()
	log.Println("Production/development database seeding completed successfully")
}
