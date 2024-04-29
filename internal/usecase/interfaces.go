package usecase

import "github.com/Leopold1975/yadro_app/internal/models"

type Storage interface {
	AddOne(ci models.ComicsInfo)
	GetByID(id string) (models.ComicsInfo, error)
	GetByWord(word string, resultLen int) []models.ComicsInfo
	Flush(updateIndex bool) (int, int, error)
}
