package notifications

type NotificationType string

const (
	NotificationTypeSystem  NotificationType = "system"
	NotificationTypeMessage NotificationType = "message"
	// NotificationTypeInvite   NotificationType = "invite" // this is correct or not?
	// NotificationTypeAlert    NotificationType = "alert"  // this is correct or not?
	NotificationTypeDocument NotificationType = "document"
	NotificationTypeESign    NotificationType = "esign"
	NotificationTypePayment  NotificationType = "payment"
)

func (n NotificationType) String() string {
	return string(n)
}

func (n NotificationType) IsValid() bool {
	return n == NotificationTypeSystem ||
		n == NotificationTypeMessage ||
		// n == NotificationTypeInvite ||
		// n == NotificationTypeAlert ||
		n == NotificationTypeDocument ||
		n == NotificationTypeESign ||
		n == NotificationTypePayment
}
