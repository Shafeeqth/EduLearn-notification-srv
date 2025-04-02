package inapp

import "go.uber.org/zap"

type InAppSender struct {
	logger *zap.Logger
	// needed to add database dependencies
}

func NewInAppSender(logger *zap.Logger) *InAppSender {
	return &InAppSender{logger: logger}
}

func (i *InAppSender) SendInApp(userId, message string) error {

	// need to add logic to store to db
	i.logger.Info("In-app notification send",
		zap.String("userId", userId),
		zap.String("message", message))

	return nil
}
