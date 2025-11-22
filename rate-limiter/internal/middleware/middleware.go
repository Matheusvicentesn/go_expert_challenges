package middleware

import (
	"net"
	"net/http"
	"rate-limiter/internal/limiter"
	"strings"
)

type RateLimitMiddleware struct {
	limiter *limiter.RateLimiter
}

func NewRateLimitMiddleware(l *limiter.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: l}
}

func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}

		token := r.Header.Get("API_KEY")

		allowed, err := m.limiter.Check(ctx, ip, token)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
