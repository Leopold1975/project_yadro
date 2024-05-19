package middlewares

import (
	"net/http"

	"github.com/Leopold1975/yadro_app/internal/pkg/config"
)

type Concurrencylimiter struct {
	limiter chan struct{}
}

func NewConcurrencylimiter(cfg config.APIConcurrency) Concurrencylimiter {
	return Concurrencylimiter{
		limiter: make(chan struct{}, cfg),
	}
}

func (cl Concurrencylimiter) ConcurrencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case cl.limiter <- struct{}{}:
			defer func() { <-cl.limiter }()

			next.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}
	})
}

func (cl Concurrencylimiter) Close() {
	close(cl.limiter)
}
