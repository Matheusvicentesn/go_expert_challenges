package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStrategy struct {
	client *redis.Client
}

func NewRedisStrategy(client *redis.Client) *RedisStrategy {
	return &RedisStrategy{client: client}
}

var limiterScript = redis.NewScript(`
	local key_count = KEYS[1]
	local key_block = KEYS[2]
	local limit = tonumber(ARGV[1])
	local block_time = tonumber(ARGV[2])

	if redis.call("EXISTS", key_block) == 1 then
		return 0
	end

	local current = redis.call("INCR", key_count)

	if current == 1 then
		redis.call("EXPIRE", key_count, 1)
	end

	if current > limit then
		redis.call("SET", key_block, "1")
		redis.call("EXPIRE", key_block, block_time)
		return 0
	end

	return 1
`)

func (r *RedisStrategy) IsAllowed(ctx context.Context, key string, limit int64, blockDuration time.Duration) (bool, error) {
	keyCount := fmt.Sprintf("limiter:count:%s", key)
	keyBlock := fmt.Sprintf("limiter:block:%s", key)

	result, err := limiterScript.Run(ctx, r.client, []string{keyCount, keyBlock}, limit, int(blockDuration.Seconds())).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
