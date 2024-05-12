package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/Leopold1975/yadro_app/internal/models"
	"github.com/Leopold1975/yadro_app/pkg/logger"
	"github.com/Leopold1975/yadro_app/pkg/words"
)

const (
	ResultLen = 10 // количетсво возвращаемых комиксов.
)

type FindComicsUsecase struct {
	db Storage
	l  logger.Logger
}

func NewComicsFind(db Storage, l logger.Logger) FindComicsUsecase {
	return FindComicsUsecase{
		db: db,
		l:  l,
	}
}

func (f FindComicsUsecase) GetComics(ctx context.Context, phrase string) ([]models.ComicsInfo, error) {
	ids, err := f.GetIDs(ctx, phrase)
	if err != nil {
		return nil, err
	}

	result := make([]models.ComicsInfo, 0, len(ids))

	for _, id := range ids {
		c, e := f.db.GetByID(ctx, id)
		if e != nil {
			err = errors.Join(err, e)
		}

		result = append(result, c)
	}

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, models.ErrNotFound
	}

	return result, nil
}

func (f FindComicsUsecase) GetIDs(ctx context.Context, phrase string) ([]string, error) {
	normalizedPhrase, err := words.StemWords(phrase)
	if err != nil {
		return nil, fmt.Errorf("stem words error: %w", err)
	}

	ids := make([][]string, 0, len(normalizedPhrase))

	for _, word := range normalizedPhrase {
		wordIDs := make([]string, 0, ResultLen)

		comics, err := f.db.GetByWord(ctx, word, ResultLen)
		if err != nil {
			f.l.Error("get by word", "error", err)
		}

		for _, c := range comics {
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
			break
		}
	}

	return result
}
