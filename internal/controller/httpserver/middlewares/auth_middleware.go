package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Leopold1975/yadro_app/internal/auth/models"
	"github.com/Leopold1975/yadro_app/internal/auth/usecase"
)

func AuthMidleware(next http.Handler, auth usecase.AuthUserUsecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)

			return
		}

		tokenBearer := r.Header.Get("Authorization")

		if tokenBearer == "" {
			http.Error(w, "missing Authorization Header", http.StatusUnauthorized)

			return
		}

		s := strings.Split(tokenBearer, " ")
		if len(s) < 2 { //nolint:gomnd
			http.Error(w, "invalid bearer token", http.StatusBadRequest)

			return
		}

		token := s[1]

		role, err := auth.Auth(r.Context(), token)
		if err != nil {
			http.Error(w, fmt.Errorf("auth error %w", err).Error(), http.StatusBadRequest)
		}

		ctxR := context.WithValue(r.Context(), models.RoleKey, models.Role(role))
		r = r.WithContext(ctxR)

		next.ServeHTTP(w, r)
	})
}
