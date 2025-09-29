package notifications

import (
	"context"
	"fmt"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultNotificationPreferences defines the default notification preferences for new users
type DefaultNotificationPreferences struct {
	NotificationType string
	Enabled          bool
	DeliveryMethod   string
}

// GetDefaultNotificationPreferences returns the default notification preferences for new users
func GetDefaultNotificationPreferences() []DefaultNotificationPreferences {
	return []DefaultNotificationPreferences{
		{
			NotificationType: "message",
			Enabled:          true,
			DeliveryMethod:   "all",
		},
		{
			NotificationType: "document",
			Enabled:          true,
			DeliveryMethod:   "all",
		},
		{
			NotificationType: "esign",
			Enabled:          true,
			DeliveryMethod:   "all",
		},
	}
}

// CreateDefaultNotificationPreferences creates default notification preferences for a new user
func CreateDefaultNotificationPreferences(ctx context.Context, queries *db.Queries, userID uuid.UUID, logger *zap.Logger) error {
	defaultPrefs := GetDefaultNotificationPreferences()

	for _, pref := range defaultPrefs {
		_, err := queries.CreateNotificationPreference(ctx, db.CreateNotificationPreferenceParams{
			UserID:           userID,
			NotificationType: pref.NotificationType,
			Enabled:          &pref.Enabled,
			DeliveryMethod:   &pref.DeliveryMethod,
		})

		if err != nil {
			logger.Error("Failed to create default notification preference",
				zap.String("userID", userID.String()),
				zap.String("notificationType", pref.NotificationType),
				zap.Error(err))
			return fmt.Errorf("failed to create notification preference for type %s: %w", pref.NotificationType, err)
		}

		logger.Debug("Created default notification preference",
			zap.String("userID", userID.String()),
			zap.String("notificationType", pref.NotificationType),
			zap.Bool("enabled", pref.Enabled),
			zap.String("deliveryMethod", pref.DeliveryMethod))
	}

	logger.Info("Successfully created default notification preferences for new user",
		zap.String("userID", userID.String()),
		zap.Int("preferencesCount", len(defaultPrefs)))

	return nil
}

// CreateDefaultNotificationPreferencesForCustomer creates default notification preferences for a new customer user
// with customer-specific defaults (more email-focused for external users)
func CreateDefaultNotificationPreferencesForCustomer(ctx context.Context, queries *db.Queries, userID uuid.UUID, logger *zap.Logger) error {
	// Customer users get more email-focused defaults since they're external users
	customerPrefs := []DefaultNotificationPreferences{
		{
			NotificationType: "message",
			Enabled:          true,
			DeliveryMethod:   "all",
		},
		{
			NotificationType: "document",
			Enabled:          true,
			DeliveryMethod:   "all",
		},
		{
			NotificationType: "esign",
			Enabled:          true,
			DeliveryMethod:   "all",
		},
	}

	for _, pref := range customerPrefs {
		_, err := queries.CreateNotificationPreference(ctx, db.CreateNotificationPreferenceParams{
			UserID:           userID,
			NotificationType: pref.NotificationType,
			Enabled:          &pref.Enabled,
			DeliveryMethod:   &pref.DeliveryMethod,
		})

		if err != nil {
			logger.Error("Failed to create default notification preference for customer",
				zap.String("userID", userID.String()),
				zap.String("notificationType", pref.NotificationType),
				zap.Error(err))
			return fmt.Errorf("failed to create notification preference for type %s: %w", pref.NotificationType, err)
		}

		logger.Debug("Created default notification preference for customer",
			zap.String("userID", userID.String()),
			zap.String("notificationType", pref.NotificationType),
			zap.Bool("enabled", pref.Enabled),
			zap.String("deliveryMethod", pref.DeliveryMethod))
	}

	logger.Info("Successfully created default notification preferences for new customer user",
		zap.String("userID", userID.String()),
		zap.Int("preferencesCount", len(customerPrefs)))

	return nil
}
