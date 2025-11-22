package storage

import (
	"context"
	"time"
)

type StorageStrategy interface {
	IsAllowed(ctx context.Context, key string, limit int64, blockDuration time.Duration) (bool, error)
}
