package models

import (
	"fmt"
	"strconv"

	"github.com/Leopold1975/yadro_app/pkg/words"
)

type ComicsInfo struct {
	ID       string   `json:"id"`
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

var ErrNotFound = fmt.Errorf("resource not found") //nolint:perfsprint

type XKCDModel struct {
	Title      string `json:"safe_title"` //nolint:tagliatelle
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Num        int    `json:"num"`
}

func ToDBComicsInfo(m XKCDModel) (ComicsInfo, error) {
	keywords, err := words.StemWords(m.Alt + " " + m.Transcript + " " + m.Title)
	if err != nil {
		return ComicsInfo{}, fmt.Errorf("stem words error: %w", err)
	}

	return ComicsInfo{
		ID:       strconv.Itoa(m.Num),
		URL:      m.Img,
		Keywords: keywords,
	}, nil
}
