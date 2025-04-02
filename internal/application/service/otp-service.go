package service

import (
	"context"
	_ "fmt"
	"os"
	"strings"
	"time"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"go.uber.org/zap"
)

type OTPService struct {
	notificationService *NotificationService
	otpRepo             domain.OTPRepository
	logger              *zap.Logger
}

func NewOTPService(notificationService *NotificationService, otpRepo domain.OTPRepository, logger *zap.Logger) *OTPService {
	return &OTPService{
		notificationService: notificationService,
		otpRepo:             otpRepo,
		logger:              logger,
	}

}

func (s *OTPService) SendOTP(ctx context.Context, userId, email, username string) (string, error) {

	// Generate OTP
	code, err := domain.GenerateOTP()
	if err != nil {
		s.logger.Error("Failed to generate OTP", zap.Error(err))
		return "", err
	}

	// Save OTP with expiration
	otp := domain.OTP{
		Code:      code,
		UserId:    userId,
		Email:     email,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minutes expiration time
	}

	if err := s.otpRepo.SaveOTP(ctx, otp); err != nil {
		s.logger.Error("Failed to save OTP", zap.Error(err))
		return "", err
	}

	subject := "Your OTP for Email Verification"
	// Load HTML template from file
	templatePath := ".\\internal\\shared\\template\\activation-mail.html"
	htmlTemplate, err := os.ReadFile(templatePath)
	if err != nil {
		s.logger.Error("Failed to read OTP HTML template", zap.Error(err))
		return "", err
	}

	// Replace placeholders in the template with actual values
	body := string(htmlTemplate)
	body = strings.ReplaceAll(body, "{{USER_NAME}}", username)
	body = strings.ReplaceAll(body, "{{CODE}}", code)
	body = strings.ReplaceAll(body, "{{EXPIRY_TIME}}", "10 minutes") // Replace with the expiration time

	// Ensure the body contains valid HTML content
	if err := s.notificationService.SendEmailNotification(ctx, userId, email, subject, body); err != nil {
		s.logger.Error("Failed to send OTP email", zap.Error(err))
		return "", err
	}

	s.logger.Info("OTP sent", zap.String("email", email), zap.String("userId", userId))
	return code, nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, userId, email, otp string) (bool, error) {
	// Retrieve OTP from repository
	storedOTP, err := s.otpRepo.GetOTP(ctx, email)
	if err != nil {
		s.logger.Error("Failed to retrieve OTP", zap.Error(err))
		return false, err
	}

	// Check if OTP matches and is not expired
	if storedOTP.Code != otp || time.Now().After(storedOTP.ExpiresAt) {
		s.logger.Warn("Invalid or expired OTP", zap.String("user_id", userId), zap.String("email", email))
		return false, nil
	}

	// OTP is valid
	s.logger.Info("OTP verified successfully", zap.String("user_id", userId), zap.String("email", email))
	return true, nil
}
