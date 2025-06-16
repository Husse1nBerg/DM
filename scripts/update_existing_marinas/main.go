package main

import (
	"context"
	"log"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/models"
)

// UpdateMarinaModules updates all marinas in the database with default modules
func UpdateMarinaModules() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.DB)
	q := db.Queries()

	// Get all marinas from database
	marinas, err := q.GetAllMarinas(ctx)
	if err != nil {
		log.Fatalf("failed to get marinas: %v", err)
	}

	if len(marinas) == 0 {
		log.Println("No marinas found in database")
		return
	}

	log.Printf("Found %d marinas to update", len(marinas))

	// Create default modules
	defaultModules := models.DefaultModules()
	defaultModulesBytes, err := defaultModules.ToBytes()
	if err != nil {
		log.Fatalf("failed to convert default modules to bytes: %v", err)
	}
	notesMessagesPlan, err := q.GetNotesMessagesPlanByName(ctx, "Free")
	if err != nil {
		log.Fatalf("failed to get notes/messages plan: %v", err)
	}
	storagePlan, err := q.GetStoragePlanByName(ctx, "Free")
	if err != nil {
		log.Fatalf("failed to get storage plan: %v", err)
	}

	// Update each marina with default modules
	updatedCount := 0
	for _, marina := range marinas {
		_, err := q.UpdateMarina(ctx, sqlc.UpdateMarinaParams{
			ID:                  marina.ID,
			Name:                marina.Name,
			Email:               marina.Email,
			Location:            marina.Location,
			Phone:               marina.Phone,
			Country:             marina.Country,
			Currency:            marina.Currency,
			WorkingHours:        marina.WorkingHours,
			Website:             marina.Website,
			Image:               marina.Image,
			MaxUsers:            marina.MaxUsers,
			IsActive:            marina.IsActive,
			IsTest:              marina.IsTest,
			AddressID:           marina.AddressID,
			SystemID:            marina.SystemID,
			Modules:             defaultModulesBytes,
			NotesMessagesPlanID: notesMessagesPlan.ID,
			StoragePlanID:       storagePlan.ID,
		})
		if err != nil {
			log.Printf("failed to update marina %s (ID: %d): %v", marina.Name, marina.ID, err)
			continue
		}
		log.Printf("Updated marina: %s (ID: %d)", marina.Name, marina.ID)
		updatedCount++
	}

	log.Printf("Successfully updated %d out of %d marinas with default modules", updatedCount, len(marinas))
	log.Printf("Default modules applied: %s", defaultModules.String())
}

func main() {
	log.Println("Starting marina modules update...")
	UpdateMarinaModules()
	log.Println("Marina modules update completed")
}
