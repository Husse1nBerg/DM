package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/models"
	u "github.com/dockworks/dm-web-backend/pkg/utils"
)

func main() {
	log.Println("Starting role permissions update...")
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.DB)
	q := db.Queries()

	roles := []struct {
		name           string
		description    string
		permissions    models.Permissions
		isActive       bool
		isCustomerRole bool
		roleType       string
		marinaID       uuid.UUID
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
				"roles.create":          true,
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
				// E-sign module objects
				"esign_templates.read":     true,
				"esign_templates.create":   true,
				"esign_templates.write":    true,
				"esign_templates.delete":   true,
				"esign_documents.read":     true,
				"esign_documents.create":   true,
				"esign_documents.write":    true,
				"esign_documents.delete":   true,
				"esign_submissions.read":   true,
				"esign_submissions.create": true,
				"esign_submissions.write":  true,
				"esign_submissions.delete": true,
				// Admin module objects
				"admin.read":   true,
				"admin.write":  true,
				"admin.delete": true,
				"admin.create": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "internal",
			marinaID:       uuid.Nil,
		},
		{
			name:        "superuser_viewer",
			description: "Superuser viewer",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":        true,
				"profile.write":       true,
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
				// E-sign module objects
				"esign_templates.read":     true,
				"esign_templates.create":   false,
				"esign_templates.write":    false,
				"esign_templates.delete":   false,
				"esign_documents.read":     true,
				"esign_documents.create":   false,
				"esign_documents.write":    false,
				"esign_documents.delete":   false,
				"esign_submissions.read":   true,
				"esign_submissions.create": false,
				"esign_submissions.write":  false,
				"esign_submissions.delete": false,
				// Admin module objects
				"admin.read": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "internal",
			marinaID:       uuid.Nil,
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
				"roles.write":           true,
				"roles.delete":          true,
				"roles.create":          true,
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
				// E-sign module objects
				"esign_templates.read":     true,
				"esign_templates.create":   true,
				"esign_templates.write":    true,
				"esign_templates.delete":   true,
				"esign_documents.read":     true,
				"esign_documents.create":   true,
				"esign_documents.write":    true,
				"esign_documents.delete":   true,
				"esign_submissions.read":   true,
				"esign_submissions.create": true,
				"esign_submissions.write":  true,
				"esign_submissions.delete": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
			marinaID:       uuid.Nil,
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
				"roles.write":           true,
				"roles.delete":          true,
				"roles.create":          true,
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
				// E-sign module objects
				"esign_templates.read":     true,
				"esign_templates.create":   true,
				"esign_templates.write":    true,
				"esign_templates.delete":   true,
				"esign_documents.read":     true,
				"esign_documents.create":   true,
				"esign_documents.write":    true,
				"esign_documents.delete":   true,
				"esign_submissions.read":   true,
				"esign_submissions.create": true,
				"esign_submissions.write":  true,
				"esign_submissions.delete": true,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
			marinaID:       uuid.Nil,
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
				// E-sign module objects
				"esign_templates.read":     true,
				"esign_templates.create":   true,
				"esign_templates.write":    true,
				"esign_templates.delete":   true,
				"esign_documents.read":     true,
				"esign_documents.create":   true,
				"esign_documents.write":    true,
				"esign_documents.delete":   false,
				"esign_submissions.read":   true,
				"esign_submissions.create": true,
				"esign_submissions.write":  true,
				"esign_submissions.delete": false,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
			marinaID:       uuid.Nil,
		},
		{
			name:        "viewer",
			description: "Read-only access",
			permissions: models.Permissions{
				// Core module objects
				"profile.read":        true,
				"profile.write":       true,
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
				// E-sign module objects
				"esign_templates.read":     true,
				"esign_templates.create":   false,
				"esign_templates.write":    false,
				"esign_templates.delete":   false,
				"esign_documents.read":     true,
				"esign_documents.create":   false,
				"esign_documents.write":    false,
				"esign_documents.delete":   false,
				"esign_submissions.read":   true,
				"esign_submissions.create": false,
				"esign_submissions.write":  false,
				"esign_submissions.delete": false,
			},
			isActive:       true,
			isCustomerRole: false,
			roleType:       "marina",
			marinaID:       uuid.Nil,
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
				// E-sign module objects
				"esign_templates.read":     false,
				"esign_templates.create":   false,
				"esign_templates.write":    false,
				"esign_templates.delete":   false,
				"esign_documents.read":     false,
				"esign_documents.create":   false,
				"esign_documents.write":    false,
				"esign_documents.delete":   false,
				"esign_submissions.read":   true,
				"esign_submissions.create": false,
				"esign_submissions.write":  false,
				"esign_submissions.delete": false,
				// Contacts module objects
				"contacts.read":   true,
				"contacts.write":  false,
				"contacts.delete": false,
			},
			isActive:       true,
			isCustomerRole: true,
			roleType:       "customer",
			marinaID:       uuid.Nil,
		},
	}

	for _, r := range roles {
		permBytes, err := r.permissions.ToBytes()
		if err != nil {
			log.Printf("failed to convert permissions to bytes for role %s: %v", r.name, err)
			continue
		}

		existingRole, err := q.GetRoleByName(ctx, r.name)
		if err != nil {
			if err == pgx.ErrNoRows {
				log.Printf("role %s does not exist, skipping", r.name)
				continue
			} else {
				log.Printf("failed to get role %s: %v", r.name, err)
				continue
			}
		}

		_, err = q.UpdateRole(ctx, sqlc.UpdateRoleParams{
			ID:             existingRole.ID,
			Name:           r.name,
			Description:    u.Pointer(r.description),
			Permissions:    permBytes,
			IsActive:       u.Pointer(r.isActive),
			IsCustomerRole: u.Pointer(r.isCustomerRole),
			Type:           r.roleType,
			Column8:        r.marinaID,
		})
		if err != nil {
			log.Printf("failed to update role %s: %v", r.name, err)
			continue
		}
		log.Printf("Updated role: %s with new permission structure", r.name)
	}

	log.Println("Role permissions update completed.")
}
