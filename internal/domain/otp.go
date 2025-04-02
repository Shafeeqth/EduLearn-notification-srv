package domain

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

type OTP struct {
	Code      string
	Email     string
	UserId    string
	ExpiresAt time.Time
}

type OTPRepository interface {
	SaveOTP(ctx context.Context, otp OTP) error
	GetOTP(ctx context.Context, email string) (OTP, error)
}

func GenerateOTP() (string, error) {
	const otpLength = 6
	const digits = "0123456789"
	code := make([]byte, otpLength)
	for i := 0; i < otpLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}

		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}
