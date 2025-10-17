package main

import (
	"context"
	"log"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
)

// SaveMarinaUsageHistory saves current marina usage data to history and resets current values
func SaveMarinaUsageHistory() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.DB)
	q := db.Queries()

	// Get all marinas
	marinas, err := q.GetAllMarinas(ctx)
	if err != nil {
		log.Fatalf("Failed to get marinas: %v", err)
	}

	savedCount := 0
	for _, marina := range marinas {
		// Skip if marina has no usage data
		if marina.StorageUsage == nil && marina.EmailUsage == nil && marina.TextUsage == nil && marina.DocumentUsage == nil && marina.AiFormDetectionUsage == nil && marina.AiComposeMessageUsage == nil {
			continue
		}

		// Get current usage values, defaulting to 0 if nil
		storageUsage := int64(0)
		emailUsage := int16(0)
		textUsage := int16(0)
		documentUsage := int64(0)
		aiFormDetectionUsage := int16(0)
		aiComposeMessageUsage := int16(0)

		if marina.StorageUsage != nil {
			storageUsage = *marina.StorageUsage
		}
		if marina.EmailUsage != nil {
			emailUsage = *marina.EmailUsage
		}
		if marina.TextUsage != nil {
			textUsage = *marina.TextUsage
		}
		if marina.DocumentUsage != nil {
			documentUsage = *marina.DocumentUsage
		}
		if marina.AiFormDetectionUsage != nil {
			aiFormDetectionUsage = *marina.AiFormDetectionUsage
		}
		if marina.AiComposeMessageUsage != nil {
			aiComposeMessageUsage = *marina.AiComposeMessageUsage
		}

		// Create usage history record
		params := sqlc.CreateMarinaUsageHistoryParams{
			MarinaID:              marina.ID,
			StorageUsage:          storageUsage,
			EmailUsage:            emailUsage,
			TextUsage:             textUsage,
			DocumentUsage:         &documentUsage,
			AiFormDetectionUsage:  &aiFormDetectionUsage,
			AiComposeMessageUsage: &aiComposeMessageUsage,
		}

		_, err := q.CreateMarinaUsageHistory(ctx, params)
		if err != nil {
			log.Printf("Failed to save usage history for marina %s: %v", marina.Name, err)
			continue
		}

		// Reset current usage values to 0
		zeroEmail := int16(0)
		zeroText := int16(0)
		zeroDocument := int64(0)
		zeroAIFormDetection := int16(0)
		zeroAIComposeMessage := int16(0)

		updateParams := sqlc.UpdateMarinaParams{
			ID:                    marina.ID,
			Name:                  marina.Name,
			Email:                 marina.Email,
			Location:              marina.Location,
			Phone:                 marina.Phone,
			Country:               marina.Country,
			Currency:              marina.Currency,
			WorkingHours:          marina.WorkingHours,
			Website:               marina.Website,
			Image:                 marina.Image,
			MaxUsers:              marina.MaxUsers,
			IsActive:              marina.IsActive,
			IsTest:                marina.IsTest,
			AddressID:             marina.AddressID,
			SystemID:              marina.SystemID,
			Modules:               marina.Modules,
			StoragePlanID:         marina.StoragePlanID,
			NotesMessagesPlanID:   marina.NotesMessagesPlanID,
			EmailUsage:            &zeroEmail,
			TextUsage:             &zeroText,
			DocumentPlanID:        marina.DocumentPlanID,
			DocumentUsage:         &zeroDocument,
			AiFormDetectionUsage:  &zeroAIFormDetection,
			AiComposeMessageUsage: &zeroAIComposeMessage,
		}

		_, err = q.UpdateMarina(ctx, updateParams)
		if err != nil {
			log.Printf("Failed to reset usage values for marina %s: %v", marina.Name, err)
			continue
		}

		savedCount++
		log.Printf("Saved usage history and reset values for marina: %s (Storage: %d bytes, Email: %d, Text: %d, Document: %d, AI Form Detection: %d, AI Compose Message: %d)",
			marina.Name, storageUsage, emailUsage, textUsage, documentUsage, aiFormDetectionUsage, aiComposeMessageUsage)
	}

	log.Printf("Usage history save complete. Processed %d marinas.", savedCount)
}

func main() {
	log.Println("Starting marina usage history save...")
	SaveMarinaUsageHistory()
	log.Println("Marina usage history save completed successfully")
}
