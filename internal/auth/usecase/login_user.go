package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leopold1975/yadro_app/internal/auth/models"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/internal/pkg/jwtauth"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserUsecase struct {
	db  Storage
	cfg config.Auth
}

func NewLoginUser(cfg config.Auth, db Storage) LoginUserUsecase {
	return LoginUserUsecase{
		db:  db,
		cfg: cfg,
	}
}

func (a LoginUserUsecase) Login(ctx context.Context, username, password string) (string, error) {
	u, err := a.db.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get user error %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return "", models.ErrWrongPassword
		}

		return "", fmt.Errorf("compare password error: %w", err)
	}

	t, err := jwtauth.GetToken(u, a.cfg.TokenMaxTime, a.cfg.Secret)
	if err != nil {
		return "", fmt.Errorf("get token error: %w", err)
	}

	return t, nil
}
