package notification

import (
	"context"
	"fmt"

	"github.com/Shafeeqth/notification-service/internal/domain"
)

// implement strategy pattern to abstract notification channels
// in order to avoid adding and removing of different messaging
// strategies

type SenderStrategy interface {
	Send(ctx context.Context, notification domain.Notification) error
}

type NotificationSender struct {
	strategies map[domain.NotificationType]SenderStrategy
}

func NewNotificationSender(strategies map[domain.NotificationType]SenderStrategy) *NotificationSender {
	return &NotificationSender{strategies: strategies}
}

func (s *NotificationSender) Send(ctx context.Context, notification domain.Notification) error {
	strategy, exists := s.strategies[notification.Type]

	if !exists {
		return fmt.Errorf("no strategy found for notification type: %s", notification.Type)
	}
	return strategy.Send(ctx, notification)
}
