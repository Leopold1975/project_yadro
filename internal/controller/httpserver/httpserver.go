package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Leopold1975/yadro_app/internal/pkg/config"
)

type Server struct {
	s *http.Server
}

func New(cfg config.Server, h http.Handler) Server {
	s := &http.Server{ //nolint:exhaustruct
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      h,
	}

	return Server{s}
}

func (s Server) Start() error {
	if err := s.s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and server error: %w", err)
	}

	return nil
}

func (s Server) Stop(ctx context.Context) error {
	if err := s.s.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	return nil
}
