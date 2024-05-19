package middlewares

import (
	"net/http"

	"github.com/Leopold1975/yadro_app/internal/pkg/config"
)

type ConcurrencyLimitter struct {
	limitter chan struct{}
}

func NewConcurrencyLimitter(cfg config.APIConcurrency) ConcurrencyLimitter {
	return ConcurrencyLimitter{
		limitter: make(chan struct{}, cfg),
	}
}

func (cl ConcurrencyLimitter) ConcurrencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case cl.limitter <- struct{}{}:
			defer func() { <-cl.limitter }()

			next.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}
	})
}

func (cl ConcurrencyLimitter) Close() {
	close(cl.limitter)
}
