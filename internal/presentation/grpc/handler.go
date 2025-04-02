package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Shafeeqth/notification-service/internal/application/service"
	"github.com/Shafeeqth/notification-service/internal/infrastructure/metrics"
	"github.com/Shafeeqth/notification-service/internal/proto"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Helper function to check if a string is alphanumeric
var isAlnum = func(s string) bool {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

type Handler struct {
	proto.UnimplementedNotificationServiceServer
	notificationService *service.NotificationService
	otpService          *service.OTPService
	logger              *zap.Logger
	validator           *validator.Validate
}

func NewHandler(notificationService *service.NotificationService, otpService *service.OTPService, logger *zap.Logger) *Handler {
	v := validator.New()

	// Register custom validation for username
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		// Custom logic: username must be alphanumeric and at least 3 characters long
		return len(username) >= 3 && isAlnum(username)
	})

	// Register custom validation for email
	v.RegisterValidation("email", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		// Custom logic: email must contain "@" and "."
		return len(email) > 3 && strings.Contains(email, "@") && strings.Contains(email, ".")
	})

	return &Handler{
		notificationService: notificationService,
		otpService:          otpService,
		logger:              logger,
		validator:           v,
	}
}

func (h *Handler) SendOTP(ctx context.Context, req *proto.OTPRequest) (*proto.NotificationResponse, error) {
	// Validate request using validator
	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: "Invalid request data"}, nil
	}

	h.logger.Info("Request received to handler (:)")
	_, err := h.otpService.SendOTP(ctx, req.UserId, req.Email, req.Username)
	if err != nil {
		h.logger.Error("Failed to send OTP", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: err.Error()}, nil
	}

	metrics.OTPSentTotal.Inc()
	return &proto.NotificationResponse{Success: true, Message: "OTP sent successfully"}, nil
}

func (h *Handler) VerifyOTP(ctx context.Context, req *proto.VerifyOTPRequest) (*proto.NotificationResponse, error) {
	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: "Invalid request data"}, nil
	}

	// Verify OTP
	isValid, err := h.otpService.VerifyOTP(ctx, req.UserId, req.Email, req.Otp)
	if err != nil {
		h.logger.Error("Failed to verify OTP", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: err.Error()}, nil
	}

	if !isValid {
		return &proto.NotificationResponse{Success: false, Message: "Invalid OTP"}, nil
	}

	h.logger.Info("OTP verified successfully", zap.String("user_id", req.UserId))
	return &proto.NotificationResponse{Success: true, Message: "OTP verified successfully"}, nil
}

func (h *Handler) ForgotPassword(ctx context.Context, req *proto.ForgotPasswordRequest) (*proto.NotificationResponse, error) {
	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: "Invalid request data"}, nil
	}

	// Send password reset email
	subject := "Password Reset Request"
	body := fmt.Sprintf("Click the following link to reset your password: %s", req.ResetLink)
	if err := h.notificationService.SendEmailNotification(ctx, req.UserId, req.Email, subject, body); err != nil {
		h.logger.Error("Failed to send password reset email", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: err.Error()}, nil
	}

	h.logger.Info("Password reset email sent", zap.String("email", req.Email))
	return &proto.NotificationResponse{Success: true, Message: "Password reset email sent successfully"}, nil
}

func (h *Handler) GetANotification(ctx context.Context, req *proto.GetNotificationRequest) (*proto.Notification, error) {
	// Validate request using validator
	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid request data")
	}

	notification, err := h.notificationService.GetANotification(ctx, req.NotificationId, req.UserId)
	if err != nil {
		h.logger.Error("Failed to get notification",
			zap.String("userId", req.UserId),
			zap.String("notification_id", req.NotificationId),
			zap.Error(err))
		return nil, err
	}
	return &proto.Notification{
			Id:        notification.ID,
			UserId:    notification.UserId,
			Type:      string(notification.Type),
			Subject:   notification.Subject,
			Body:      notification.Body,
			Recipient: notification.Recipient,
			IsRead:    notification.IsRead,
			CreatedAt: notification.CreatedAt.Format(time.RFC3339),
		},
		nil
}

func (h *Handler) GetAllNotifications(ctx context.Context, req *proto.GetAllNotificationsRequest) (*proto.GetAllNotificationsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}

	pageSize := int(req.PageSize)
	if pageSize < 1 {
		page = 10
	}
	var isRead *bool
	if req.IsRead {
		isRead = &req.IsRead
	}
	var notifyType *string
	if req.Type != "" {
		notifyType = &req.Type
	}

	notifications, total, err := h.notificationService.GetAllNotifications(ctx, req.UserId, page, pageSize, isRead, notifyType)
	if err != nil {
		h.logger.Error("Failed to get notifications", zap.String("userId", req.UserId), zap.Error(err))
		return nil, err
	}
	protoNotifications := make([]*proto.Notification, len(notifications))
	for i, n := range notifications {
		protoNotifications[i] = &proto.Notification{
			Id:        n.ID,
			UserId:    n.UserId,
			Type:      string(n.Type),
			Subject:   n.Subject,
			Body:      n.Body,
			Recipient: n.Recipient,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Format(time.RFC3339),
		}
	}
	return &proto.GetAllNotificationsResponse{
		Notifications: protoNotifications,
		Total:         int32(total),
		Page:          int32(page),
		PageSize:      int32(pageSize),
	}, nil
}

func (h *Handler) MarkAsRead(ctx context.Context, req *proto.MarkNotificationRequest) (*proto.NotificationResponse, error) {
	if err := h.notificationService.MarkAsRead(ctx, req.NotificationId, req.UserId); err != nil {
		h.logger.Error("Failed to mark as read", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: err.Error()}, nil

	}
	return &proto.NotificationResponse{Success: true, Message: "Notification marked as read"}, nil
}

func (h *Handler) MarkAllAsRead(ctx context.Context, req *proto.MarkAllNotificationsRequest) (*proto.NotificationResponse, error) {
	if err := h.notificationService.MarkAllAsRead(ctx, req.UserId); err != nil {
		h.logger.Error("Failed to mark all as read", zap.Error(err))
		return &proto.NotificationResponse{Success: false, Message: err.Error()}, nil
	}
	return &proto.NotificationResponse{Success: true, Message: "All notifications marked as read"}, nil
}
