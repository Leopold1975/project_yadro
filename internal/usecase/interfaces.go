package usecase

import (
	"context"

	"github.com/Leopold1975/yadro_app/internal/models"
)

type Storage interface {
	AddOne(ctx context.Context, ci models.ComicsInfo) error
	GetByID(ctx context.Context, id string) (models.ComicsInfo, error)
	GetByWord(ctx context.Context, word string, resultLen int) ([]models.ComicsInfo, error)
	Flush(ctx context.Context, updateIndex bool) (int, int, error)
}
