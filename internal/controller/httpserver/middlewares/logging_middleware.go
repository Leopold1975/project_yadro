package middlewares

import (
	"net/http"
	"time"

	"github.com/Leopold1975/yadro_app/pkg/logger"
)

// Q: Как лучше, использовать "перехватчик", перезагружая 1 из методов,
// или использовать готовый httptest.ResponseRecorder?

// Полностью копирует исходный ResponseWriter, за исключением метода
// WriteHeader: метод также перехватывает StatusCode.
type syncResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *syncResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.ResponseWriter.WriteHeader(code)

	rw.status = code
	rw.wroteHeader = true
}

func LogMiddleware(next http.Handler, l logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := syncResponseWriter{ResponseWriter: w} //nolint:exhaustruct

		defer func() {
			if rr.status == 0 {
				rr.status = http.StatusOK // HTTP 200 OK, если статус код не был установлен.
			}

			latency := time.Since(start)
			l.Info(r.Proto,
				"METHOD", r.Method,
				"Addr", r.URL.RequestURI(),
				"Client", r.RemoteAddr,
				"Agent", r.UserAgent(),
				"CODE", rr.status,
				"LATENCY", latency.String(),
			)
		}()

		next.ServeHTTP(&rr, r)
	})
}
