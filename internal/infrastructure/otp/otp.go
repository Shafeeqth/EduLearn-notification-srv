package otp

import (
	"fmt"
	"sync"
	"time"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"go.uber.org/zap"
)

type OTPRepository struct {
	store  map[string]domain.OTP
	mutex  sync.RWMutex
	logger *zap.Logger
}

func NewOTPRepository(logger *zap.Logger) *OTPRepository {
	return &OTPRepository{
		store:  make(map[string]domain.OTP),
		logger: logger,
	}
}

func (r *OTPRepository) SaveOTP(otp domain.OTP) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.store[otp.Email] = otp
	r.logger.Info("OTP saved", zap.String("email", otp.Email))
	return nil
}

func (r *OTPRepository) GetOTP(email string) (domain.OTP, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	otp, exists := r.store[email]
	if !exists {
		r.logger.Warn("OTP not found", zap.String("email", email))
		return domain.OTP{}, fmt.Errorf("OTP not found for email: %s", email)
	}

	if time.Now().After(otp.ExpiresAt) {
		r.logger.Warn("OTP expired", zap.String("email", email))
		return domain.OTP{}, fmt.Errorf("OTP expired for email: %s", email)
	}

	return otp, nil
}
