package xkcd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Leopold1975/yadro_app/pkg/database"
	"github.com/Leopold1975/yadro_app/pkg/words"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnexpectedCode = errors.New("unexpected response from server")
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
