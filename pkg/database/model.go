package database

import "errors"

var ErrNotFound = errors.New("resource not found")

type ComicsInfo struct {
	ID       string   `json:"-"`
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}
