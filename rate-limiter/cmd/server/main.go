package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
	"rate-limiter/internal/middleware"
	"rate-limiter/internal/storage"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	redisStrategy := storage.NewRedisStrategy(rdb)

	limiterConfig := limiter.Config{
		RateLimitIP:    cfg.RateLimitIP,
		RateLimitToken: cfg.RateLimitToken,
		BlockTime:      time.Duration(cfg.BlockTime) * time.Second,
	}

	rateLimiter := limiter.NewRateLimiter(redisStrategy, limiterConfig)

	rlMiddleware := middleware.NewRateLimitMiddleware(rateLimiter)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Request Allowed."))
	})

	handler := rlMiddleware.Handler(mux)

	serverAddr := fmt.Sprintf(":%s", cfg.ServerPort)
	fmt.Printf("Server running on port %s\n", serverAddr)
	fmt.Printf("Limits -> IP: %d req/s | Token: %d req/s | Block: %ds\n",
		cfg.RateLimitIP, cfg.RateLimitToken, cfg.BlockTime)

	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}
