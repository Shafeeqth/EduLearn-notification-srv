package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EduLearn/notification-service/internal/domain"
	"github.com/EduLearn/notification-service/internal/infrastructure/email"
	"github.com/EduLearn/notification-service/internal/infrastructure/redis"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository struct {
	db    *pgxpool.Pool
	email *email.EmailSender
	// optStore map[string]domain.OTP
	redis    *redis.RedisClient
	otpMutex sync.RWMutex
	logger   *zap.Logger
}

func NewRepository(db *DB, emailSender *email.EmailSender, redisClient *redis.RedisClient, logger *zap.Logger) *Repository {
	return &Repository{
		db:    db.pool,
		email: emailSender,
		redis: redisClient,
		// optStore: make(map[string]domain.OTP),
		logger: logger,
	}
}

func (r *Repository) SaveNotification(ctx context.Context, notification domain.Notification) error {
	query := `
	INSERT INTO notification (id, user_id, type, subject, body, is_ready, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, notification.ID, notification.UserId, notification.Type, notification.Subject, notification.Body, notification.IsRead, notification.CreatedAt)
	if err != nil {
		r.logger.Error("Failed to save notifications", zap.Error(err))
		return err
	}
	r.logger.Info("Notification saved", zap.String("id", notification.ID))
	return nil
}

func (r *Repository) BatchSaveNotifications(ctx context.Context, notifications []domain.Notification) error {
	batch := &pgx.Batch{}
	for _, n := range notifications {
		query := `
		INSERT INTO notifications (id, use_id, type, subject, body, is_read, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		batch.Queue(query, n.ID, n.UserId, n.Type, n.Subject, n.Body, n.IsRead, n.CreatedAt)

	}
	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		r.logger.Error("Failed to batch save notifications", zap.Error(err))
		return err
	}
	r.logger.Info("Batch saved notifications", zap.Int("count", len(notifications)))
	return nil

}

func (r *Repository) GetAllNotifications(ctx context.Context, userId string) ([]domain.Notification, error) {
	query := `
	SELECT id, user_id, type, subject, body, is_read, created_at 
	From notifications WHERE user_id = $1 
	ORDER BY created_at DESC
	LIMIT 100
	`
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		r.logger.Error("Failed to get notifications", zap.String("user_id", userId))
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserId, &n.Subject, &n.Type, &n.Subject, &n.Body, &n.IsRead, &n.CreatedAt); err != nil {
			r.logger.Error("Failed to scan notification", zap.Error(err))
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (r *Repository) CheckIfProcessed(ctx context.Context, notificationID string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM processed_notifications WHERE notification_id = $1)`
    err := r.db.QueryRow(ctx, query, notificationID).Scan(&exists)
    if err != nil {
        r.logger.Error("Failed to check if notification processed",
            zap.String("notification_id", notificationID),
            zap.Error(err))
        return false, domain.ErrDatabase
    }
    return exists, nil
}

func (r *Repository) MarkAsProcessed(ctx context.Context, notificationID string) error {
    query := `INSERT INTO processed_notifications (notification_id) VALUES ($1) ON CONFLICT DO NOTHING`
    _, err := r.db.Exec(ctx, query, notificationID)
    if err != nil {
        r.logger.Error("Failed to mark notification as processed",
            zap.String("notification_id", notificationID),
            zap.Error(err))
        return domain.ErrDatabase
    }
    return nil
}

func (r *Repository) MarkAsRead(ctx context.Context, notificationId, userId string) error {
	query := `
	UPDATE notifications 
	SET is_read = TRUE
	WHERE id= $1 and user_id = $2 
	`
	result, err := r.db.Exec(ctx, query, notificationId, userId)
	if err != nil {
		r.logger.Error("Failed to mark notification as read", zap.Error(err))
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found or not authorized")
	}
	return nil
}

func (r *Repository) MarkAllAsRead(ctx context.Context, userId string) error {
	query := `
	UPDATE notifications 
	SET is_read = TRUE
	WHERE and user_id = $1 
	`
	result, err := r.db.Exec(ctx, query, userId)
	if err != nil {
		r.logger.Error("Failed to mark all notifications as read", zap.Error(err))
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("notifications not found or not authorized")
	}
	return nil
}

func (r *Repository) SendEmail(ctx context.Context, recipient, subject, body string) error {
	return r.email.SendEmail(ctx, recipient, subject, body)
}

func (r *Repository) SaveOTP(ctx context.Context, otp domain.OTP) error {
	r.otpMutex.Lock()
	defer r.otpMutex.Unlock()

	r.redis.SaveOTP(ctx, otp)
	r.logger.Info("OTP saved", zap.String("email", otp.Email))
	return nil
}

func (r *Repository) GetOTP(ctx context.Context, email string) (domain.OTP, error) {
	r.otpMutex.RLock()
	defer r.otpMutex.RUnlock()

	otp, err := r.redis.GetOTP(ctx, email)
	if err != nil {
		r.logger.Warn("OTP not found", zap.String("email", email))
		return domain.OTP{}, fmt.Errorf("OTP not found for  n : %s", email)
	}

	if time.Now().After(otp.ExpiresAt) {
		r.logger.Warn("OTP expired", zap.String("email", email))
		return domain.OTP{}, fmt.Errorf("OTP expired for email: %s", email)

	}
	return otp, nil

}
