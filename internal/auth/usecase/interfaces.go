package usecase

import (
	"context"

	"github.com/Leopold1975/yadro_app/internal/auth/models"
)

type Storage interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUser(ctx context.Context, username string) (models.User, error)
}
