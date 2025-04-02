package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationService struct {
	repo   domain.NotificationRepository
	logger *zap.Logger
	kafka  KafkaProducer
}

type KafkaProducer interface {
	Produce(ctx context.Context, topic string, message []byte) error
}

func NewNotificationService(repo domain.NotificationRepository, logger *zap.Logger, kafka KafkaProducer) *NotificationService {

	return &NotificationService{repo: repo, logger: logger, kafka: kafka}
}

// func (s *NotificationService) SendEmailNotification(ctx context.Context, userId, recipient, subject, body string) error {
// 	notification := domain.Notification{
// 		ID:        uuid.New().String(),
// 		UserId:    userId,
// 		Subject:   subject,
// 		Body:      body,
// 		Type:      domain.EmailNotification,
// 		IsRead:    false,
// 		CreatedAt: time.Now(),
// 	}

// 	// save to database
// 	if err := s.repo.SaveNotification(ctx, notification); err != nil {
// 		s.logger.Error("Failed to save notification to db", zap.Error(err))

// 	}

// 	// Send to Kafka for asynchronous processing
// 	msg, err := json.Marshal(notification)
// 	if err != nil {
// 		s.logger.Error("Failed to marshall email notification", zap.Error(err))
// 		return err
// 	}

// 	if err := s.kafka.Produce(ctx, string(util.EmailNotifications), msg); err != nil {
// 		s.logger.Error("Failed to produce email notification to Kafka", zap.Error(err))
// 		return err
// 	}
// 	s.logger.Info("email notification queued", zap.String("recipient", recipient))
// 	return nil
// }

func (s *NotificationService) SendInAppNotification(ctx context.Context, userId, recipient, subject, body string, notifyType domain.NotificationType) error {
	notification := domain.Notification{
		ID:        uuid.New().String(),
		UserId:    userId,
		Subject:   subject,
		Type:      notifyType,
		Body:      body,
		Recipient: recipient,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// save to database
	if err := s.repo.SaveNotification(ctx, notification); err != nil {
		s.logger.Error("Failed to save notification to db", zap.Error(err))

	}

	s.logger.Info("In-app notification queued", zap.String("recipient", userId))
	return nil

}
func (s *NotificationService) SendNotification(ctx context.Context, userId, recipient, subject, body string, notifyType domain.NotificationType) error {
	notification := domain.Notification{
		ID:        uuid.New().String(),
		UserId:    uuid.New().String(),// userId,
		Subject:   subject,
		Type:      notifyType,
		Body:      body,
		Recipient: recipient,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// save to database
	if err := s.repo.SaveNotification(ctx, notification); err != nil {
		s.logger.Error("Failed to save notification to db", zap.Error(err))

	}

	msg, err := json.Marshal(notification)
	if err != nil {
		s.logger.Error("Failed to marshall in-app notification", zap.Error(err))
		return err
	}

	topic := notifyType + "-notifications"
	if err := s.kafka.Produce(ctx, string(topic), msg); err != nil {
		s.logger.Error("Failed to produce in-app notification to Kafka", zap.Error(err))
		return domain.ErrKafkaProduce
	}
	s.logger.Info("Notification queued",
		zap.String("recipient", recipient),
		zap.String("userId", userId),
		zap.String("type", string(notifyType)))
	return nil

}

func (s *NotificationService) GetANotification(ctx context.Context, notificationId, userId string) (*domain.Notification, error) {
	notification, err := s.repo.GetANotification(ctx, notificationId, userId)
	if err != nil {
		s.logger.Error("Failed to get notification",
			zap.String("notification_id", notificationId),
			zap.String("userId", userId),
			zap.Error(err))
		return nil, err
	}
	return notification, nil
}

func (s *NotificationService) GetAllNotifications(ctx context.Context, userId string, page, pageSize int, isRead *bool, notifyType *string) ([]domain.Notification, int64, error) {
	notifications, total, err := s.repo.GetAllNotifications(ctx, userId, page, pageSize, isRead, notifyType)
	if err != nil {
		s.logger.Error("Failed to get notifications from db", zap.String("userId", userId), zap.Error(err))
		return nil, 0, err
	}
	return notifications, total, nil
}

func (s *NotificationService) SendEmailNotification(ctx context.Context, userId, recipient, subject, body string) error {
	return s.SendNotification(ctx, userId, recipient, subject, body, domain.EmailNotification)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, notificationId, userId string) error {
	if err := s.repo.MarkAsRead(ctx, notificationId, userId); err != nil {
		s.logger.Error("Failed to mark notification as read",
			zap.String("notificationId",
				notificationId), zap.String("userId", userId),
			zap.Error(err))
		return err
	}
	s.logger.Info("Notification marked as read", zap.String("notificationId", notificationId), zap.String("userId", userId))
	return nil
}
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userId string) error {
	if err := s.repo.MarkAllAsRead(ctx, userId); err != nil {
		s.logger.Error("Failed to mark notifications all as read",
			zap.String("userId", userId),
			zap.Error(err))
		return err
	}
	s.logger.Info("Notification marked all as read",
		zap.String("userId", userId))
	return nil
}
