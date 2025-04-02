package ratelimit

import (
	"context"
	"sync"

	"github.com/Shafeeqth/notification-service/internal/domain"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.Mutex
	rate     float64 // Requests per second
	burst    int     // Burst size
}

func NewRateLimiter(rates float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rates,
		burst:    burst,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, key string) error {
	r.mutex.Lock()
	limiter, exists := r.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(r.rate), r.burst)
		r.limiters[key] = limiter
	}
	r.mutex.Unlock()
	if err := limiter.Wait(ctx); err != nil {
		return domain.ErrRateLimit
	}
	return nil

}
