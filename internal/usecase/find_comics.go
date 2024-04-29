package usecase

import (
	"fmt"
	"sort"

	"github.com/Leopold1975/yadro_app/internal/models"
	"github.com/Leopold1975/yadro_app/pkg/words"
)

const (
	ResultLen = 10
)

type FindComicsUsecase struct {
	db Storage
}

func NewComicsFind(db Storage) FindComicsUsecase {
	return FindComicsUsecase{
		db: db,
	}
}

func (f FindComicsUsecase) GetComics(phrase string) ([]models.ComicsInfo, error) {
	ids, err := f.GetIDs(phrase)
	if err != nil {
		return nil, err
	}

	result := make([]models.ComicsInfo, 0, len(ids))

	for _, id := range ids {
		c, err := f.db.GetByID(id)
		if err != nil {
			continue
		}

		result = append(result, c)
	}

	return result, nil
}

func (f FindComicsUsecase) GetIDs(phrase string) ([]string, error) {
	normalizedPhrase, err := words.StemWords(phrase)
	if err != nil {
		return nil, fmt.Errorf("stem words error: %w", err)
	}

	ids := make([][]string, 0, len(normalizedPhrase))

	for _, word := range normalizedPhrase {
		wordIDs := make([]string, 0, ResultLen)
		for _, c := range f.db.GetByWord(word, ResultLen) {
			wordIDs = append(wordIDs, c.ID)
		}

		ids = append(ids, wordIDs)
	}

	result := GetTopIDs(ids...)

	return result, nil
}

// GetTopIDs получает пересечение переданных слайсов с учетом частоты,
// с которой элементы пересечения встречаются.
func GetTopIDs(results ...[]string) []string {
	result := make([]string, 0, ResultLen)
	h := make(map[string]int)

	for _, res := range results {
		for _, word := range res {
			h[word]++
		}
	}

	sortedIDs := make([]struct {
		Key   string
		Value int
	}, 0, len(h))

	for k, v := range h {
		sortedIDs = append(sortedIDs, struct {
			Key   string
			Value int
		}{k, v})
	}

	sort.Slice(sortedIDs, func(i, j int) bool {
		return sortedIDs[i].Value > sortedIDs[j].Value
	})

	for i := 0; i < len(sortedIDs); i++ {
		result = append(result, sortedIDs[i].Key)
		if len(result) == ResultLen {
			return result
		}
	}

	return result
}
