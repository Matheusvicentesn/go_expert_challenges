package limiter_test

import (
	"context"
	"rate-limiter/internal/limiter"
	"testing"
	"time"
)

type MockStrategy struct {
	counts map[string]int
	blocks map[string]bool
}

func (m *MockStrategy) IsAllowed(ctx context.Context, key string, limit int64, blockDuration time.Duration) (bool, error) {
	if m.blocks[key] {
		return false, nil
	}

	m.counts[key]++

	if int64(m.counts[key]) > limit {
		m.blocks[key] = true
		return false, nil
	}
	return true, nil
}

func TestRateLimiter_IP(t *testing.T) {
	mock := &MockStrategy{counts: make(map[string]int), blocks: make(map[string]bool)}
	cfg := limiter.Config{RateLimitIP: 2, RateLimitToken: 5, BlockTime: time.Minute}
	rl := limiter.NewRateLimiter(mock, cfg)
	ctx := context.Background()

	if allowed, _ := rl.Check(ctx, "192.168.0.1", ""); !allowed {
		t.Error("Esperado permitido, recebeu bloqueado")
	}

	if allowed, _ := rl.Check(ctx, "192.168.0.1", ""); !allowed {
		t.Error("Esperado permitido, recebeu bloqueado")
	}

	if allowed, _ := rl.Check(ctx, "192.168.0.1", ""); allowed {
		t.Error("Esperado bloqueado, recebeu permitido")
	}
}

func TestRateLimiter_TokenOverride(t *testing.T) {
	mock := &MockStrategy{counts: make(map[string]int), blocks: make(map[string]bool)}
	cfg := limiter.Config{RateLimitIP: 1, RateLimitToken: 3, BlockTime: time.Minute}
	rl := limiter.NewRateLimiter(mock, cfg)
	ctx := context.Background()

	token := "abc-123"

	rl.Check(ctx, "10.0.0.1", token)
	rl.Check(ctx, "10.0.0.1", token)
	allowed, _ := rl.Check(ctx, "10.0.0.1", token)

	if !allowed {
		t.Error("Token deveria permitir a 3ª requisição")
	}

	allowed, _ = rl.Check(ctx, "10.0.0.1", token)
	if allowed {
		t.Error("Token deveria bloquear a 4ª requisição")
	}
}
