package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shafeeqth/notification-service/internal/application/service"
	"github.com/Shafeeqth/notification-service/internal/domain"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/config"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/database"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/email"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/kafka"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/logging"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/metrics"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/notification"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/redis"
	"github.com/Shafeeqth/notification-service/internal/presentation/grpc"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Import the PostgreSQL driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type NotificationRepository struct {
	repo *database.Repository
	// emailSender *email.EmailSender
	// inAppSender *inapp.InAppSender
}

/*
type NotificationRepository interface {
	SaveNotification(notification Notification) error
	GetAllNotifications(userId string) ([]Notification, error)
	MarkAsRead(notificationId, userId string) error
	MarkAllAsRead(userId string) error
	SendEmail(recipient, subject, body string) error
	// SendInApp(userId, message string) error*/

func (r *NotificationRepository) SaveNotification(ctx context.Context, notification domain.Notification) error {
	return r.repo.SaveNotification(ctx, notification)
}
func (r *NotificationRepository) GetANotification(ctx context.Context, notificationId, userId string) (*domain.Notification, error) {
	return r.repo.GetANotification(ctx, notificationId, userId)
}
func (r *NotificationRepository) GetAllNotifications(ctx context.Context, userId string, page, pageSize int, isRead *bool, notifyType *string) ([]domain.Notification, int64, error) {
	return r.repo.GetAllNotifications(ctx, userId, page, pageSize, isRead, notifyType)
}
func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationId, userId string) error {
	return r.repo.MarkAsRead(ctx, notificationId, userId)
}
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userId string) error {
	return r.repo.MarkAllAsRead(ctx, userId)
}
func (r *NotificationRepository) MarkAsProcessed(ctx context.Context, notificationId string) error {
	return r.repo.MarkAsProcessed(ctx, notificationId)
}
func (r *NotificationRepository) CheckIfProcessed(ctx context.Context, notificationId string) (bool, error) {
	return r.repo.CheckIfProcessed(ctx, notificationId)
}

func (r *NotificationRepository) SendEmail(ctx context.Context, recipient, subject, body string) error {
	return r.repo.SendEmail(ctx, recipient, subject, body)

}

func (r *NotificationRepository) SaveOTP(ctx context.Context, otp domain.OTP) error {
	return r.repo.SaveOTP(ctx, otp)
}

func (r *NotificationRepository) GetOTP(ctx context.Context, email string) (domain.OTP, error) {

	return r.repo.GetOTP(ctx, email)
}

func (r *NotificationRepository) AutoMigrate() error {
	return r.repo.AutoMigrate()
}

// func (r *NotificationRepository) SendEmail(recipient, subject, body string) error {
// 	return r.SendEmail(recipient, subject, body)
// }

// func (r *NotificationRepository) SendInApp(userID, message string) error {
// 	return r.inAppSender.SendInApp(userID, message)
// }

func main() {

	// initialize logger
	logger, err := logging.NewLogger()
	if err != nil {
		panic("Failed to initialize logger" + err.Error())
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Fatal("Fatal to load config", zap.Error(err))

	}

	// Initialize metrics and server
	metrics.InitMetrics()
	metrics.StartMetricsServer()

	// // Run database migrations
	// m, err := migrate.New("file://migrations", cfg.DatabaseDSN)
	// if err != nil {
	// 	logger.Fatal("Failed to initialize migration", zap.Error(err))
	// }

	// // Check if the database is in a dirty state
	// if _, _, databaseErr := m.Version(); databaseErr != nil {
	// 	var migrateErr *migrate.ErrDirty
	// 	if errors.As(databaseErr, &migrateErr) {
	// 		logger.Error("Database is in a dirty state", zap.Error(databaseErr))
	// 		// Force the migration version to the last successful version
	// 		if err := m.Force(2025032401); err != nil {
	// 			logger.Fatal("Failed to force migration version", zap.Error(err))
	// 		}
	// 		logger.Info("Forced migration version to 2025032401")
	// 	} else {
	// 		logger.Fatal("Failed to get migration version", zap.Error(databaseErr))
	// 	}
	// }

	// // Apply migrations
	// if err := m.Up(); err != nil && err != migrate.ErrNoChange {
	// 	logger.Error("Failed to apply migrations", zap.Error(err))
	// } else {
	// 	logger.Info("Database migrations applied")
	// }

	// Initialize database
	db, err := database.NewDB(cfg.DatabaseDSN, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))

	}
	defer db.Close()

	redisClient, err := redis.NewRedisClient(cfg.RedisAddr, logger)
	if err != nil {
		logger.Fatal("Failed to initialize redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer
	KafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka producer", zap.Error(err))
	}

	defer KafkaProducer.Close()

	// initialize notification  senders
	emailSender, err := email.NewEmailSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, logger)
	// inAppSender := inapp.NewInAppSender(logger)
	if err != nil {
		logger.Fatal("Failed to initialize sender", zap.Error(err))
	}

	// implement all strategies here
	strategies := map[domain.NotificationType]notification.SenderStrategy{
		domain.EmailNotification: emailSender,
	}
	notificationSender := notification.NewNotificationSender(strategies)

	// initialize repository
	repo := database.NewRepository(db, notificationSender, redisClient, logger)
	notificationRepo := &NotificationRepository{repo: repo}

	// Auto migrate database schema
	if err := notificationRepo.AutoMigrate(); err != nil {
		logger.Error("Failed to run migration", zap.Error(err))

	}

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.ConsumerGroup, notificationRepo, notificationSender, logger, 5, 3)
	if err != nil {
		logger.Fatal("Failed to initialize Kafka consumer", zap.Error(err))
	}
	defer consumer.Close()

	// Start kafka consumers
	go func() {
		if err := consumer.ConsumeNotifications(); err != nil {
			logger.Fatal("Failed to consume email notifications", zap.Error(err))
		}
	}()

	// go func() {
	// 	if err := consumer.ConsumeInAppNotifications(); err != nil {
	// 		logger.Fatal("Failed to consume in-app notifications", zap.Error(err))
	// 	}
	// }()

	// initialize services
	notificationService := service.NewNotificationService(notificationRepo, logger, KafkaProducer)
	// otpRepo := otp.NewOTPRepository(logger)
	otpService := service.NewOTPService(notificationService, notificationRepo, logger)

	// Start grpc Server
	grpcServer := grpc.NewServer(notificationService, otpService, logger)

	go func() {
		if err := grpcServer.Start(":" + string(cfg.GRpcPort)); err != nil {
			logger.Fatal("Failed to start rRPC server", zap.Error(err))
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
	grpcServer.Stop()
}
