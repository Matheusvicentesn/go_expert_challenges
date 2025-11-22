package limiter

import (
	"context"
	"rate-limiter/internal/storage"
	"time"
)

type Config struct {
	RateLimitIP    int64
	RateLimitToken int64
	BlockTime      time.Duration
}

type RateLimiter struct {
	strategy storage.StorageStrategy
	config   Config
}

func NewRateLimiter(strategy storage.StorageStrategy, config Config) *RateLimiter {
	return &RateLimiter{
		strategy: strategy,
		config:   config,
	}
}

func (rl *RateLimiter) Check(ctx context.Context, ip string, token string) (bool, error) {
	var key string
	var limit int64

	if token != "" {
		key = "token:" + token
		limit = rl.config.RateLimitToken
	} else {
		key = "ip:" + ip
		limit = rl.config.RateLimitIP
	}

	return rl.strategy.IsAllowed(ctx, key, limit, rl.config.BlockTime)
}
