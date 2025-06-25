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

			// Create marina modules using the new model
			marinaModules := models.DefaultModules()
			marinaModulesBytes, err := marinaModules.ToBytes()
			if err != nil {
				log.Fatalf("failed to create marina modules: %v", err)
			}

			notesMessagesPlan, err := q.GetNotesMessagesPlanByName(ctx, "Free")
			if err != nil {
				log.Fatalf("failed to get notes/messages plan: %v", err)
			}
			storagePlan, err := q.GetStoragePlanByName(ctx, "Free")
			if err != nil {
				log.Fatalf("failed to get storage plan: %v", err)
			}
			marina, err = q.CreateMarina(ctx, sqlc.CreateMarinaParams{
				Name:                "Dockmaster Web",
				Email:               "marina@dockmaster.com",
				IsActive:            u.Pointer(true),
				AddressID:           marinaAddress.ID,
				OrganizationID:      org.ID,
				WorkingHours:        workingHoursBytes,
				Modules:             marinaModulesBytes,
				NotesMessagesPlanID: notesMessagesPlan.ID,
				StoragePlanID:       storagePlan.ID,
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
		name           string
		description    string
		permissions    models.Permissions
		isActive       bool
		isCustomerRole bool
		roleType       string
	}{
		{
			name:        "superuser",
			description: "Platform Superuser with full system access",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":          true,
				"profile.write":         true,
				"profile.delete":        true,
				"users.read":            true,
				"users.write":           true,
				"users.delete":          true,
				"organizations.read":    true,
				"organizations.write":   true,
				"organizations.delete":  true,
				"marinas.read":          true,
				"marinas.write":         true,
				"marinas.delete":        true,
				"addresses.read":        true,
				"addresses.write":       true,
				"addresses.delete":      true,
				"roles.read":            true,
				"roles.write":           true,
				"roles.delete":          true,
				"marina_gallery.read":   true,
				"marina_gallery.write":  true,
				"marina_gallery.delete": true,
				// Customer & Vessels module objects
				"customers.read":      true,
				"customers.write":     true,
				"customers.delete":    true,
				"customers.create":    true,
				"vessels.read":        true,
				"vessels.write":       true,
				"vessels.delete":      true,
				"vessels.create":      true,
				"messages.read":       true,
				"messages.write":      true,
				"messages.delete":     true,
				"messages.create":     true,
				"documents.read":      true,
				"documents.write":     true,
				"documents.delete":    true,
				"documents.create":    true,
				"boat_gallery.read":   true,
				"boat_gallery.write":  true,
				"boat_gallery.delete": true,
				"boat_gallery.create": true,
				// Service Management module objects
				"work_orders.read":   true,
				"work_orders.write":  true,
				"work_orders.delete": true,
				"work_orders.create": true,
				// History module objects
				"history.read": true,
				// Plans module objects
				"plans.read": true,
				// Settings module objects
				"settings.read":  true,
				"settings.write": true,
				// Contacts module objects
				"contacts.read":   true,
				"contacts.write":  true,
				"contacts.delete": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "internal",
		},
		{
			name:        "org_admin",
			description: "Organization Administrator",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":          true,
				"profile.write":         true,
				"users.read":            true,
				"users.write":           true,
				"users.delete":          true,
				"users.create":          true,
				"organizations.read":    true,
				"organizations.write":   true,
				"organizations.delete":  false,
				"marinas.read":          true,
				"marinas.write":         true,
				"marinas.delete":        false,
				"marinas.create":        false,
				"addresses.read":        true,
				"addresses.write":       true,
				"addresses.delete":      true,
				"addresses.create":      true,
				"roles.read":            true,
				"roles.write":           false,
				"roles.delete":          false,
				"marina_gallery.read":   true,
				"marina_gallery.write":  true,
				"marina_gallery.delete": true,
				"marina_gallery.create": true,
				// Customer & Vessels module objects
				"customers.read":      true,
				"customers.write":     true,
				"customers.delete":    true,
				"customers.create":    true,
				"vessels.read":        true,
				"vessels.write":       true,
				"vessels.delete":      true,
				"vessels.create":      true,
				"messages.read":       true,
				"messages.write":      true,
				"messages.delete":     true,
				"messages.create":     true,
				"documents.read":      true,
				"documents.write":     true,
				"documents.delete":    true,
				"documents.create":    true,
				"boat_gallery.read":   true,
				"boat_gallery.write":  true,
				"boat_gallery.delete": true,
				"boat_gallery.create": true,
				// Service Management module objects
				"work_orders.read":   true,
				"work_orders.write":  true,
				"work_orders.delete": true,
				"work_orders.create": true,
				// History module objects
				"history.read": true,
				// Plans module objects
				"plans.read": true,
				// Settings module objects
				"settings.read":  true,
				"settings.write": true,
				// Contacts module objects
				"contacts.read":   true,
				"contacts.write":  true,
				"contacts.delete": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
		},
		{
			name:        "marina_admin",
			description: "Administrator at the marina level",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":          true,
				"profile.write":         true,
				"users.read":            true,
				"users.write":           true,
				"users.delete":          true,
				"users.create":          true,
				"organizations.read":    true,
				"organizations.write":   false,
				"organizations.delete":  false,
				"marinas.read":          true,
				"marinas.write":         true,
				"marinas.delete":        false,
				"addresses.read":        true,
				"addresses.write":       true,
				"addresses.delete":      false,
				"addresses.create":      true,
				"roles.read":            true,
				"roles.write":           false,
				"roles.delete":          false,
				"marina_gallery.read":   true,
				"marina_gallery.write":  true,
				"marina_gallery.delete": true,
				"marina_gallery.create": true,
				// Customer & Vessels module objects
				"customers.read":      true,
				"customers.write":     true,
				"customers.delete":    true,
				"customers.create":    true,
				"vessels.read":        true,
				"vessels.write":       true,
				"vessels.delete":      true,
				"vessels.create":      true,
				"messages.read":       true,
				"messages.write":      true,
				"messages.delete":     false,
				"messages.create":     true,
				"documents.read":      true,
				"documents.write":     true,
				"documents.delete":    true,
				"documents.create":    true,
				"boat_gallery.read":   true,
				"boat_gallery.write":  true,
				"boat_gallery.delete": true,
				"boat_gallery.create": true,
				// Service Management module objects
				"work_orders.read":   true,
				"work_orders.write":  true,
				"work_orders.delete": false,
				"work_orders.create": true,
				// History module objects
				"history.read": true,
				// Plans module objects
				"plans.read": true,
				// Settings module objects
				"settings.read":  true,
				"settings.write": true,
				// Contacts module objects
				"contacts.read":   true,
				"contacts.write":  true,
				"contacts.delete": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
		},
		{
			name:        "staff",
			description: "Regular staff member with limited write access",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":          true,
				"profile.write":         true,
				"users.create":          true,
				"users.read":            true,
				"users.write":           true,
				"users.delete":          true,
				"organizations.read":    true,
				"organizations.write":   false,
				"organizations.delete":  false,
				"marinas.read":          true,
				"marinas.write":         false,
				"marinas.delete":        false,
				"addresses.read":        true,
				"addresses.write":       false,
				"addresses.delete":      false,
				"roles.read":            true,
				"roles.write":           false,
				"roles.delete":          false,
				"marina_gallery.read":   true,
				"marina_gallery.write":  false,
				"marina_gallery.delete": false,
				// Customer & Vessels module objects
				"customers.read":      true,
				"customers.write":     true,
				"customers.delete":    false,
				"customers.create":    true,
				"vessels.read":        true,
				"vessels.write":       true,
				"vessels.delete":      false,
				"vessels.create":      true,
				"messages.read":       true,
				"messages.write":      true,
				"messages.delete":     false,
				"messages.create":     true,
				"documents.read":      true,
				"documents.write":     true,
				"documents.delete":    false,
				"documents.create":    true,
				"boat_gallery.read":   true,
				"boat_gallery.write":  true,
				"boat_gallery.delete": false,
				"boat_gallery.create": true,
				// Service Management module objects
				"work_orders.read":   true,
				"work_orders.write":  true,
				"work_orders.delete": false,
				"work_orders.create": true,
				// History module objects
				"history.read": true,
				// Plans module objects
				"plans.read": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
		},
		{
			name:        "viewer",
			description: "Read-only access",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":        true,
				"users.read":          true,
				"users.write":         true,
				"organizations.read":  true,
				"marinas.read":        true,
				"addresses.read":      true,
				"roles.read":          true,
				"marina_gallery.read": true,
				// Customer & Vessels module objects
				"customers.read":    true,
				"vessels.read":      true,
				"messages.read":     true,
				"documents.read":    true,
				"boat_gallery.read": true,
				// Service Management module objects
				"work_orders.read": true,
				// History module objects
				"history.read": true,
				// Plans module objects
				"plans.read": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
		},
		{
			name:        "customer_user",
			description: "Customer User",
			permissions: models.Permissions{
				// Core module objects - limited access
				"profile.read":        true,
				"profile.write":       true,
				"users.read":          true,
				"users.write":         true,
				"marinas.read":        true,
				"addresses.read":      true,
				"roles.read":          true,
				"marina_gallery.read": true,
				// Customer & Vessels module objects - limited access
				"customers.read":      true,
				"customers.write":     true,
				"vessels.read":        true,
				"vessels.write":       true,
				"vessels.delete":      true,
				"vessels.create":      true,
				"messages.read":       true,
				"messages.write":      true,
				"messages.create":     true,
				"documents.read":      true,
				"documents.write":     true,
				"documents.create":    true,
				"boat_gallery.read":   true,
				"boat_gallery.write":  true,
				"boat_gallery.create": true,
				// History module objects
				"history.read": true,
				// Plans module objects
				"plans.read": true,
				// Contacts module objects
				"contacts.read":   true,
				"contacts.write":  false,
				"contacts.delete": false,
			},
			isActive:       true,
			isCustomerRole: true,
			roleType:       "customer",
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
					Name:           r.name,
					Description:    u.Pointer(r.description),
					Permissions:    permBytes,
					IsActive:       u.Pointer(r.isActive),
					IsCustomerRole: u.Pointer(r.isCustomerRole),
					Type:           r.roleType,
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
				ID:             existingRole.ID,
				Name:           r.name,
				Description:    u.Pointer(r.description),
				Permissions:    permBytes,
				IsActive:       u.Pointer(r.isActive),
				IsCustomerRole: u.Pointer(r.isCustomerRole),
				Type:           r.roleType,
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
				Username:       "admin" + u.RandomString(6),
				FirstName:      "Andrew",
				LastName:       "Sameh",
				Email:          cfg.App.AdminEmail,
				PasswordHash:   u.Pointer(string(encryptedPassword)),
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

func main() {
	log.Println("Starting production/development database seeding...")
	RunProdSeed()
	log.Println("Production/development database seeding completed successfully")
}
