package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	u "github.com/dockworks/dm-web-backend/pkg/utils"
)

// intToStringPtr converts *int to *string
func intToStringPtr(i *int) *string {
	if i == nil {
		return nil
	}
	v := fmt.Sprintf("%d", *i)
	return &v
}

// RunPlansSeed seeds the database with initial plans data
func RunPlansSeed() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.DB)
	q := db.Queries()

	// Seed Notes and Messages Plans
	notesMessagesPlans := []struct {
		name          string
		monthlyPrice  float64
		textLimit     *string
		emailLimit    *string
		userLimit     *string
		isMostPopular bool
	}{
		{
			name:          "Free",
			monthlyPrice:  0.00,
			textLimit:     u.Pointer("100 Texts"),
			emailLimit:    u.Pointer("Unlimited Emails"),
			userLimit:     u.Pointer("Unlimited Users"),
			isMostPopular: false,
		},
		{
			name:          "Starter",
			monthlyPrice:  150.00,
			textLimit:     u.Pointer("2500 Texts"),
			emailLimit:    u.Pointer("Unlimited Emails"),
			userLimit:     u.Pointer("Unlimited Users"),
			isMostPopular: false,
		},
		{
			name:          "Intermediate",
			monthlyPrice:  250.00,
			textLimit:     u.Pointer("7000 Texts"),
			emailLimit:    u.Pointer("Unlimited Emails"),
			userLimit:     u.Pointer("Unlimited Users"),
			isMostPopular: true,
		},
		{
			name:          "Advanced",
			monthlyPrice:  500.00,
			textLimit:     u.Pointer("Unlimited Texts"),
			emailLimit:    u.Pointer("Unlimited Emails"),
			userLimit:     u.Pointer("Unlimited Users"),
			isMostPopular: false,
		},
		{
			name:          "Pay as You Go",
			monthlyPrice:  1.50,
			textLimit:     u.Pointer("Unlimited Texts"),
			emailLimit:    u.Pointer("Unlimited Emails"),
			userLimit:     u.Pointer("Unlimited Users"),
			isMostPopular: false,
		},
	}

	for _, p := range notesMessagesPlans {
		existingPlan, err := q.GetNotesMessagesPlanByName(ctx, p.name)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Create new plan if doesn't exist
				_, err = q.CreateNotesMessagesPlan(ctx, sqlc.CreateNotesMessagesPlanParams{
					Name:          p.name,
					MonthlyPrice:  p.monthlyPrice,
					TextLimit:     p.textLimit,
					EmailLimit:    p.emailLimit,
					UserLimit:     p.userLimit,
					IsMostPopular: u.Pointer(p.isMostPopular),
				})
				if err != nil {
					log.Fatalf("failed to create notes/messages plan %s: %v", p.name, err)
				}
				log.Printf("Created notes/messages plan: %s", p.name)
			} else {
				log.Fatalf("failed to get notes/messages plan %s: %v", p.name, err)
			}
		} else {
			// Update existing plan
			_, err = q.UpdateNotesMessagesPlan(ctx, sqlc.UpdateNotesMessagesPlanParams{
				ID:            existingPlan.ID,
				Name:          p.name,
				MonthlyPrice:  p.monthlyPrice,
				TextLimit:     p.textLimit,
				EmailLimit:    p.emailLimit,
				UserLimit:     p.userLimit,
				IsMostPopular: u.Pointer(p.isMostPopular),
			})
			if err != nil {
				log.Fatalf("failed to update notes/messages plan %s: %v", p.name, err)
			}
			log.Printf("Updated notes/messages plan: %s", p.name)
		}
	}

	// Seed Storage Plans
	storagePlans := []struct {
		name           string
		monthlyPrice   float64
		storageLimitGB *string
		userLimit      *string
		isMostPopular  bool
	}{
		{
			name:           "Free",
			monthlyPrice:   0.00,
			storageLimitGB: u.Pointer("5 GB"),
			userLimit:      u.Pointer("Unlimited Users"),
			isMostPopular:  false,
		},
		{
			name:           "Starter",
			monthlyPrice:   100.00,
			storageLimitGB: u.Pointer("100 GB"),
			userLimit:      u.Pointer("Unlimited Users"),
			isMostPopular:  false,
		},
		{
			name:           "Intermediate",
			monthlyPrice:   150.00,
			storageLimitGB: u.Pointer("250 GB"),
			userLimit:      u.Pointer("Unlimited Users"),
			isMostPopular:  true,
		},
		{
			name:           "Advanced",
			monthlyPrice:   500.00,
			storageLimitGB: u.Pointer("Unlimited Storage"),
			userLimit:      u.Pointer("Unlimited Users"),
			isMostPopular:  false,
		},
		{
			name:           "Pay as You Go",
			monthlyPrice:   1.50,
			storageLimitGB: u.Pointer("Unlimited Storage"),
			userLimit:      u.Pointer("Unlimited Users"),
			isMostPopular:  false,
		},
	}

	for _, p := range storagePlans {
		existingPlan, err := q.GetStoragePlanByName(ctx, p.name)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Create new plan if doesn't exist
				_, err = q.CreateStoragePlan(ctx, sqlc.CreateStoragePlanParams{
					Name:           p.name,
					MonthlyPrice:   p.monthlyPrice,
					StorageLimitGb: p.storageLimitGB,
					UserLimit:      p.userLimit,
					IsMostPopular:  u.Pointer(p.isMostPopular),
				})
				if err != nil {
					log.Fatalf("failed to create storage plan %s: %v", p.name, err)
				}
				log.Printf("Created storage plan: %s", p.name)
			} else {
				log.Fatalf("failed to get storage plan %s: %v", p.name, err)
			}
		} else {
			// Update existing plan
			_, err = q.UpdateStoragePlan(ctx, sqlc.UpdateStoragePlanParams{
				ID:             existingPlan.ID,
				Name:           p.name,
				MonthlyPrice:   p.monthlyPrice,
				StorageLimitGb: p.storageLimitGB,
				UserLimit:      p.userLimit,
				IsMostPopular:  u.Pointer(p.isMostPopular),
			})
			if err != nil {
				log.Fatalf("failed to update storage plan %s: %v", p.name, err)
			}
			log.Printf("Updated storage plan: %s", p.name)
		}
	}
}

func main() {
	RunPlansSeed()
}
