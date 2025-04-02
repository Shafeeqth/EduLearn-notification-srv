package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/notification"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	repo          domain.NotificationRepository
	sender        *notification.NotificationSender
	logger        *zap.Logger
	workers       int
	retries       int
}

func NewConsumer(brokers []string, groupId string, repo domain.NotificationRepository, sender *notification.NotificationSender, logger *zap.Logger, workers, retries int) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		logger.Error("Failed to create Kafka consumer group", zap.Error(err))
		return nil, err
	}
	return &Consumer{consumerGroup: consumerGroup, repo: repo, logger: logger, sender: sender, workers: workers, retries: retries}, nil

}

type ConsumerHandler struct {
	repo    domain.NotificationRepository
	sender  *notification.NotificationSender
	logger  *zap.Logger
	jobs    chan *sarama.ConsumerMessage
	wg      sync.WaitGroup
	workers int
	retries int
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for i := 0; i < h.workers; i++ {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			for msg := range h.jobs {
				var notification domain.Notification
				if err := json.Unmarshal(msg.Value, &notification); err != nil {
					h.logger.Error("Failed to un-marshall notification", zap.String("topic", msg.Topic), zap.Error(err))
					continue
				}

				// check for idempotency
				processed, err := h.repo.CheckIfProcessed(context.Background(), notification.ID)
				if err != nil {
					h.logger.Error("Failed to check if notification processed",
						zap.String("notification_id", notification.ID),
						zap.Error(err))
					continue
				}
				if processed {
					h.logger.Warn("Notification already processed",
						zap.String("notification_id", notification.ID),
						zap.Error(err))
					continue
				}

				// Retry logic
				for attempt := 1; attempt <= h.retries; attempt++ {
					err := h.sender.Send(context.Background(), notification)
					if err == nil {
						break
					}
					h.logger.Warn("Failed to send notification, retrying",
						zap.String("notification_id", notification.ID), zap.Int("attempt", attempt), zap.Error(err))
					if attempt == h.retries {
						h.logger.Error("Failed to send notification after retries",
							zap.String("notification_id", notification.ID),
							zap.Error(err))
						continue
					}
					time.Sleep(time.Second * time.Duration(attempt))
				}
				if err := h.repo.MarkAsProcessed(context.Background(), notification.ID); err != nil {
					h.logger.Error("Failed to mark notification as processed",
						zap.String("notification_id", notification.ID), zap.Error(err))
					continue
				}
				session.MarkMessage(msg, "")
				h.logger.Info("Notification processed",
					zap.String("topic", msg.Topic),
					zap.Int32("partition", msg.Partition),
					zap.Int64("offset", msg.Offset))
			}

		}()
	}
	for msg := range claim.Messages() {
		h.jobs <- msg
	}

	// close jobs channel and wait for workers to finish
	close(h.jobs)
	return nil

	// Feed messages to workers

	// for msg := range claim.Messages() {
	// 	var notification domain.Notification
	// 	if err := json.Unmarshal(msg.Value, &notification); err != nil {
	// 		h.logger.Error("Failed to unmarshal notification",
	// 			zap.String("topic", msg.Topic),
	// 			zap.Error(err))
	// 		continue
	// 	}
	// 	// todo
	// 	if err := h.repo.SendEmail(notification.UserId, notification.Subject, notification.Body); err != nil {
	// 		h.logger.Error("Failed to process notification",
	// 			zap.String("topic", msg.Topic),
	// 			zap.Error(err))
	// 		continue
	// 	}
	// 	session.MarkMessage(msg, "")
	// 	h.logger.Info("Notification processed",
	// 		zap.String("topic", msg.Topic),
	// 		zap.Int32("partition", msg.Partition),
	// 		zap.Int64("offset", msg.Offset))
	// }
	// return nil

}

func (c *Consumer) ConsumeNotifications() error {

	topics := []string{
		"email-notifications",
		"sms-notifications",
		"push-notifications"}
	handler := &ConsumerHandler{
		repo:    c.repo,
		logger:  c.logger,
		jobs:    make(chan *sarama.ConsumerMessage, 100),
		workers: c.workers,
		sender:  c.sender,
		retries: c.retries,
	}

	for {
		err := c.consumerGroup.Consume(context.Background(), topics, handler)
		if err != nil {
			c.logger.Error("Failed to consume email notifications", zap.Error(err))
			return err
		}
	}
	// return c.consumeTopic("email-notifications", func(n domain.Notification) error {
	// 	return c.repo.SendEmail(n.UserId, n.Subject, n.Body)
	// })
}

// func (c *Consumer) ConsumeInAppNotifications() error {
// 	return c.consumeTopic("inapp-notifications", func(n domain.Notification) error {
// 		return c.repo.SendInApp(n.UserId, n.Message)
// 	})
// }

// func (c *Consumer) consumeTopic(topic string, handler func(domain.Notification) error) error {
// 	partitions, err := c.consumer.Partitions(topic)
// 	if err != nil {
// 		c.logger.Error("Failed to start consumer for topic", zap.String("topic", topic), zap.Error(err))
// 		return err
// 	}

// 	for _, partition := range partitions {
// 		pc, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
// 		if err != nil {
// 			c.logger.Error("Failed to start consumer for partition",
// 				zap.String("topic", topic),
// 				zap.Int32("partition", partition),
// 				zap.Error(err))
// 		}

// 		go func(pc sarama.PartitionConsumer) {
// 			for msg := range pc.Messages() {
// 				var notification domain.Notification
// 				if err := json.Unmarshal(msg.Value, &notification); err != nil {
// 					c.logger.Error("Failed to unmarshall notification",
// 						zap.String("topic", topic),
// 						zap.Error(err))
// 					continue
// 				}
// 				if err := handler(notification); err != nil {
// 					c.logger.Error("Failed to process notification",
// 						zap.String("topic", topic),
// 						zap.Error(err))
// 					continue
// 				}
// 				c.logger.Info("Notification processed",
// 					zap.String("topic", topic),
// 					zap.Int32("partition", msg.Partition),
// 					zap.Int64("offset", msg.Offset))
// 			}

// 		}(pc)
// 	}
// 	return nil
// }

func (c *Consumer) Close() error {
	if err := c.consumerGroup.Close(); err != nil {
		c.logger.Error("Failed to close Kafka consumer group", zap.Error(err))
		return err
	}
	c.logger.Info("Kafka consumer group closed")
	return nil
}
