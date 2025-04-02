package domain

import "errors"

var (
	ErrOTPNotFound      = errors.New("OTP not found")
	ErrOTPExpired       = errors.New("OTP has expired")
	ErrRateLimit        = errors.New("rate limit exceeded")
	ErrDatabase         = errors.New("database error")
	ErrKafkaProduce     = errors.New("failed to produce Kafka message")
	ErrEmailSend        = errors.New("failed to send email")
	ErrAlreadyProcessed = errors.New("notification already processed")
	ErrNotFound         = errors.New("notification not found")
	ErrUnauthorized     = errors.New("unauthorized access to resource")
)
