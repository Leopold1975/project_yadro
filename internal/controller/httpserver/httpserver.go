package httpserver

import (
	"context"
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
	return fmt.Errorf("listen and server error: %w", s.s.ListenAndServe())
}

func (s Server) Stop(ctx context.Context) error {
	if err := s.s.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	return nil
}
