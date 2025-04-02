package kafka

import (
	"context"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type Producer struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
}

// NewProducer initializes a new Kafka producer.
func NewProducer(brokers []string, logger *zap.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		logger.Error("Failed to create Kafka producer", zap.Error(err))
		return nil, err
	}

	logger.Info("Kafka producer initialized successfully")
	return &Producer{
		producer: producer,
		logger:   logger,
	}, nil
}

// // SendMessage sends a message to the specified Kafka topic.
// func (p *Producer) SendMessage(topic, key, value string) error {
// 	msg := &sarama.ProducerMessage{
// 		Topic: topic,
// 		Key:   sarama.StringEncoder(key),
// 		Value: sarama.StringEncoder(value),
// 	}

// 	partition, offset, err := p.producer.SendMessage(msg)
// 	if err != nil {
// 		p.logger.Error("Failed to send message", zap.Error(err))
// 		return err
// 	}

// 	p.logger.Info("Message sent successfully",
// 		zap.String("topic", topic),
// 		zap.Int32("partition", partition),
// 		zap.Int64("offset", offset),
// 	)
// 	return nil
// }

func (p *Producer) Produce(ctx context.Context, topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}
	select {

	case <-ctx.Done():
		return ctx.Err()
	default:
		partition, offset, err := p.producer.SendMessage(msg)
		if err != nil {
			p.logger.Error("Failed to send Kafka message",
				zap.String("topic", topic),
				zap.Error(err))
			return err
		}

		p.logger.Info("Message sent to Kafka",
			zap.String("topic", topic),
			zap.Int32("partition", partition),
			zap.Int64("offset", offset))
		return nil
	}

}

// Close closes the Kafka producer.
func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		p.logger.Error("Failed to close Kafka producer", zap.Error(err))
		return err
	}
	p.logger.Info("Kafka producer closed successfully")
	return nil
}
