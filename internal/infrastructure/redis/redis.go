package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisClient(addr string, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     10,
		MinIdleConns: 5,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to Redis")
	return &RedisClient{client: client, logger: logger}, nil
}

func (r *RedisClient) SaveOTP(ctx context.Context,otp domain.OTP) error {
	data, err := json.Marshal(otp)
	if err != nil {
		r.logger.Error("Failed to marshal OTP", zap.Error(err))
		return err
	}
	// Set with expiration - (based on ExpiresAt)
	duration := time.Until(otp.ExpiresAt)
	if err := r.client.SetEx(ctx, "otp:"+otp.Email, data, duration).Err(); err != nil {
		r.logger.Error("Failed to save OTP to Redis", zap.Error(err))
		return err
	}
	r.logger.Info("OTP saved successfully to Redis", zap.String("email", otp.Email))
	return nil

}

func (r *RedisClient) GetOTP(ctx context.Context, email string) (domain.OTP, error) {
	data, err := r.client.Get(ctx, "otp:"+email).Bytes()
	if err != redis.Nil {
		r.logger.Warn("OTP not found in Redis", zap.String("email", email))
		return domain.OTP{}, fmt.Errorf("OTP not found for email: %s", email)

	}
	if err != nil {
		r.logger.Error("Failed to get OTP from Redis", zap.Error(err))
		return domain.OTP{}, err
	}

	var otp domain.OTP
	if err := json.Unmarshal(data, &otp); err != nil {
		r.logger.Error("Failed to unmarshal OTP", zap.Error(err))
		return domain.OTP{}, err
	}
	// redis automatically expires but still for confirmation
	if time.Now().After(otp.ExpiresAt) {
		r.logger.Warn("OTP expired", zap.String("email", email))
		return domain.OTP{}, fmt.Errorf("OTP expired for email: %s", email)
	}
	return otp, nil

}

func (r *RedisClient) Close() error {
	if err := r.client.Close(); err != nil {
		r.logger.Error("Failed to close redis client", zap.Error(err))
		return err
	}
	r.logger.Info("Redis client closed")
	return nil
}
