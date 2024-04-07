package xkcd

import (
	"fmt"
	"strconv"

	"github.com/Leopold1975/yadro_app/pkg/database"
	"github.com/Leopold1975/yadro_app/pkg/words"
)

var (
	ErrNotFound       = fmt.Errorf("resource not found")              //nolint:perfsprint
	ErrUnexpectedCode = fmt.Errorf("unexpected response from server") //nolint:perfsprint
)

type Model struct {
	Num        int    `json:"num"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
}

func ToDBComicsInfo(m Model) (database.ComicsInfo, error) {
	keywords, err := words.StemWords(m.Alt + " " + m.Transcript)
	if err != nil {
		return database.ComicsInfo{}, fmt.Errorf("stem words error: %w", err)
	}

	return database.ComicsInfo{
		ID:       strconv.Itoa(m.Num),
		URL:      m.Img,
		Keywords: keywords,
	}, nil
}

func ToDBComicsInfos(models []Model) ([]database.ComicsInfo, error) {
	result := make([]database.ComicsInfo, 0, len(models))
	errs := make([]error, 0)

	for _, m := range models {
		ci, err := ToDBComicsInfo(m)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		result = append(result, ci)
	}

	var err error
	for _, e := range errs {
		if err == nil {
			err = fmt.Errorf("%w", e)

			continue
		}

		err = fmt.Errorf("%w\n%w", err, e)
	}

	return result, err
}
