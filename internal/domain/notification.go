package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID        string           `gorm:"type:uuid;primaryKey"`
	UserId    string           `gorm:"type:uuid;index"`
	Type      NotificationType `gorm:"type:varchar(50)"`
	Subject   string           `gorm:"type:text;"`
	Body      string           `gorm:"type:text"`
	Recipient string           `gorm:"type:text"`
	IsRead    bool             `gorm:"default:false;index"`
	CreatedAt time.Time        `gorm:"autoCreateTime"`
}

type ProcessedNotification struct {
	NotificationId string `gorm:"type:uuid;primaryKey"`
	CreatedAt      time.Time
}

type NotificationType string

const (
	EmailNotification NotificationType = "email"
	InAppNotification NotificationType = "inapp"
	OTPNotification   NotificationType = "otp"
)

// NotificationService defines the interface for managing notifications in the system.
// It provides methods for sending notifications, retrieving notifications, and marking them as read.
// This interface acts as a middleman between the data layer and the domain layer, and is designed
// to be implemented by any service that handles notification-related operations.
type NotificationService interface {
	// SendEmailNotification sends an email notification to a specific user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - userId: The ID of the user to whom the notification is being sent.
	// - recipient: The email address of the recipient.
	// - subject: The subject of the email.
	// - body: The body content of the email.
	// Returns:
	// - An error if the operation fails.
	SendNotification(ctx context.Context, userId, recipient, subject, body string) error

	// SendInAppNotification sends an in-app notification to a specific user.
	// Parameters:
	// - userId: The ID of the user to whom the notification is being sent.
	// - message: The message content of the in-app notification.
	// Returns:
	// - An error if the operation fails.
	SendInAppNotification(ctx context.Context, userId, message string) error

	// SendOTP sends a one-time password (OTP) notification to a user's email.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - userId: The ID of the user to whom the OTP is being sent.
	// - email: The email address of the recipient.
	// Returns:
	// - A string representing the generated OTP.
	// - An error if the operation fails.
	SendOTP(ctx context.Context, userId, email string) (string, error)

	GetANotification(ctx context.Context, notificationId, userId string) (*Notification, error)

	// GetAllNotifications retrieves all notifications for a specific user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - userId: The ID of the user whose notifications are to be retrieved.
	// Returns:
	// - A slice of Notification objects.
	// - An error if the operation fails.
	GetAllNotifications(ctx context.Context, userId string, page, pageSize int, isRead *bool, notifyType *string) ([]Notification, int64, error)

	// MarkAsRead marks a specific notification as read for a user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - notificationId: The ID of the notification to be marked as read.
	// - userId: The ID of the user for whom the notification is being marked as read.
	// Returns:
	// - An error if the operation fails.
	MarkAsRead(ctx context.Context, notificationId, userId string) error

	// MarkAllAsRead marks all notifications as read for a specific user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - userId: The ID of the user whose notifications are to be marked as read.
	// Returns:
	// - An error if the operation fails.
	MarkAllAsRead(ctx context.Context, userId string) error
}

type NotificationSender interface {
	Send(ctx context.Context, notification Notification) error
}

// NotificationRepository defines the interface for managing notifications.
// It provides methods for saving, retrieving, and updating notifications,
// as well as sending email notifications.
// NotificationRepository defines the interface for managing notifications and sending email notifications.
//
// Example usage:
//
//	// Create a new notification repository instance (implementation not shown).
//	var repo NotificationRepository
//
//	// Save a notification.
//	err := repo.SaveNotification(ctx, Notification{ID: "1", UserID: "user123", Message: "Welcome!"})
//	if err != nil {
//		log.Fatalf("Failed to save notification: %v", err)
//	}
//
//	// Retrieve all notifications for a user.
//	notifications, err := repo.GetAllNotifications(ctx, "user123")
//	if err != nil {
//		log.Fatalf("Failed to retrieve notifications: %v", err)
//	}
//	for _, n := range notifications {
//		fmt.Printf("Notification: %v\n", n)
//	}
//
//	// Mark a specific notification as read.
//	err = repo.MarkAsRead(ctx, "1", "user123")
//	if err != nil {
//		log.Fatalf("Failed to mark notification as read: %v", err)
//	}
//
//	// Mark all notifications as read for a user.
//	err = repo.MarkAllAsRead(ctx, "user123")
//	if err != nil {
//		log.Fatalf("Failed to mark all notifications as read: %v", err)
//	}
//
//	// Send an email notification.
//	err = repo.SendEmail(ctx, "recipient@example.com", "Subject", "Email body content")
//	if err != nil {
//		log.Fatalf("Failed to send email: %v", err)
//	}
type NotificationRepository interface {
	// SaveNotification saves a new notification to the repository.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - notification: The notification object to be saved.
	// Returns:
	// - An error if the operation fails.
	SaveNotification(ctx context.Context, notification Notification) error

	GetANotification(ctx context.Context, notificationId, userId string) (*Notification, error)

	// GetAllNotifications retrieves all notifications for a specific user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - userId: The ID of the user whose notifications are to be retrieved.
	// Returns:
	// - A slice of Notification objects.
	// - An error if the operation fails.
	GetAllNotifications(ctx context.Context, userId string, page, pageSize int, isRead *bool, notifyType *string) ([]Notification, int64, error)
	// MarkAsRead marks a specific notification as read for a user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - notificationId: The ID of the notification to be marked as read.
	// - userId: The ID of the user for whom the notification is being marked as read.
	// Returns:
	// - An error if the operation fails.
	MarkAsRead(ctx context.Context, notificationId, userId string) error

	// MarkAllAsRead marks all notifications as read for a specific user.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - userId: The ID of the user whose notifications are to be marked as read.
	// Returns:
	// - An error if the operation fails.
	MarkAllAsRead(ctx context.Context, userId string) error

	// SendEmail sends an email notification to a recipient.
	// Parameters:
	// - ctx: The context for managing request-scoped values, deadlines, and cancellations.
	// - recipient: The email address of the recipient.
	// - subject: The subject of the email.
	// - body: The body content of the email.
	// Returns:
	// - An error if the operation fails.
	SendEmail(ctx context.Context, recipient, subject, body string) error

	CheckIfProcessed(ctx context.Context, notificationId string) (bool, error)
	MarkAsProcessed(ctx context.Context, notificationId string) error

	AutoMigrate() error
}
