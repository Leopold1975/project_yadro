package usecase

import (
	"context"
	"fmt"

	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/internal/pkg/jwtauth"
)

type AuthUserUsecase struct {
	db  Storage
	cfg config.Auth
}

func NewAuthUser(cfg config.Auth, db Storage) AuthUserUsecase {
	return AuthUserUsecase{
		db:  db,
		cfg: cfg,
	}
}

func (a AuthUserUsecase) Auth(ctx context.Context, token string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context error %w", ctx.Err())
	default:
	}

	role, err := jwtauth.ValidateTokenRole(token, a.cfg.Secret)
	if err != nil {
		return "", fmt.Errorf("validate token role error %w", err)
	}

	return role, nil
}
