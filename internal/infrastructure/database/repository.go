package database

import (
	"context"
	"time"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/notification"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/redis"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	sender *notification.NotificationSender
	redis  *redis.RedisClient
	logger *zap.Logger
}

func NewRepository(db *DB, sender *notification.NotificationSender, redisClient *redis.RedisClient, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db.DB(),
		sender: sender,
		redis:  redisClient,
		logger: logger,
	}
}

func (r *Repository) AutoMigrate() error {
	if err := r.db.AutoMigrate(&domain.Notification{}, &domain.ProcessedNotification{}); err != nil {
		r.logger.Error("Failed to auto-migrate database", zap.Error(err))
		return err
	}
	r.logger.Info("Database migration completed")
	return nil
}

func (r *Repository) SaveNotification(ctx context.Context, notification domain.Notification) error {
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	if err := r.db.WithContext(ctx).Create(&notification).Error; err != nil {
		r.logger.Error("Failed to save notification", zap.Error(err))
		return domain.ErrDatabase
	}
	r.logger.Info("Notification saved", zap.String("id", notification.ID))
	return nil
}

func (r *Repository) GetANotification(ctx context.Context, notificationID, userID string) (*domain.Notification, error) {
	var notification domain.Notification
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Warn("Notification not found",
				zap.String("notification_id", notificationID),
				zap.String("user_id", userID))
			return nil, domain.ErrNotFound
		}
		r.logger.Error("Failed to get notification",
			zap.String("notification_id", notificationID),
			zap.Error(err))
		return nil, domain.ErrDatabase
	}
	return &notification, nil
}

func (r *Repository) GetAllNotifications(ctx context.Context, userId string, page, pageSize int, isRead *bool, notifyType *string) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userId)
	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("Failed to count notifications",
			zap.String("user_id", userId),
			zap.Error(err))
		return nil, 0, domain.ErrDatabase
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications).Error; err != nil {
		r.logger.Error("Failed to get notifications",
			zap.String("user_id", userId),
			zap.Error(err))
		return nil, 0, domain.ErrDatabase
	}
	return notifications, total, nil

}

func (r *Repository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	if result.Error != nil {
		r.logger.Error("Failed to mark notification as read",
			zap.String("notification_id", notificationID),
			zap.Error(result.Error))
		return domain.ErrDatabase
	}
	if result.RowsAffected == 0 {
		r.logger.Warn("Notification not found or unauthorized",
			zap.String("notification_id", notificationID),
			zap.String("user_id", userID))
		return domain.ErrUnauthorized
	}
	return nil
}

func (r *Repository) MarkAllAsRead(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)
	if result.Error != nil {
		r.logger.Error("Failed to mark all notifications as read",
			zap.String("user_id", userID),
			zap.Error(result.Error))
		return domain.ErrDatabase
	}
	r.logger.Info("Marked all notifications as read",
		zap.String("user_id", userID),
		zap.Int64("rows_affected", result.RowsAffected))
	return nil
}

func (r *Repository) CheckIfProcessed(ctx context.Context, notificationID string) (bool, error) {
	var processed domain.ProcessedNotification
	err := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		First(&processed).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		r.logger.Error("Failed to check if notification processed",
			zap.String("notification_id", notificationID),
			zap.Error(err))
		return false, domain.ErrDatabase
	}
	return true, nil
}

func (r *Repository) MarkAsProcessed(ctx context.Context, notificationID string) error {
	processed := domain.ProcessedNotification{
		NotificationId: notificationID,
		CreatedAt:      time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(&processed).Error; err != nil {
		r.logger.Error("Failed to mark notification as processed",
			zap.String("notification_id", notificationID),
			zap.Error(err))
		return domain.ErrDatabase
	}
	return nil
}

func (r *Repository) SendEmail(ctx context.Context, recipient, subject, body string) error {
	notification := domain.Notification{
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
		Type:      domain.EmailNotification,
	}
	return r.sender.Send(ctx, notification)
}

func (r *Repository) SaveOTP(ctx context.Context, otp domain.OTP) error {
	return r.redis.SaveOTP(ctx, otp)
}

func (r *Repository) GetOTP(ctx context.Context, email string) (domain.OTP, error) {
	return r.redis.GetOTP(ctx, email)
}
