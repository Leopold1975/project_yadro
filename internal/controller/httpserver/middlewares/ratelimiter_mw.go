package middlewares

import (
	"net"
	"net/http"
	"sync"

	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rate     int
	burst    int
}

func NewRateLimiter(cfg config.Ratelimit) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		mu:       sync.Mutex{},
		rate:     cfg.Limit,
		burst:    cfg.Burst,
	}
}

func (rl *RateLimiter) RatelimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "invalid IP address", http.StatusInternalServerError)

			return
		}

		var limiter *rate.Limiter
		var exists bool

		rl.mu.Lock()
		if limiter, exists = rl.limiters[ip]; !exists {
			limiter = rate.NewLimiter(rate.Limit(rl.rate), rl.burst)
			rl.limiters[ip] = limiter
		}

		rl.mu.Unlock()

		if limiter.Allow() {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
		}
	})
}
