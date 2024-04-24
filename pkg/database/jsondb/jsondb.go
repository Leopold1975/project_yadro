package jsondb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/Leopold1975/yadro_app/pkg/app"
	"github.com/Leopold1975/yadro_app/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/database"
)

const (
	EstimatedDBSize    = 3000
	EstimatedIndexSize = 20000
)

// we use map here because it's only json db's inner
// data object.
// External clients should use only lists (slices) and models.
type (
	JSONStorage map[string]database.ComicsInfo
	JSONIndex   map[string][]string
)

type JSONDatabase struct {
	db       JSONStorage
	index    JSONIndex
	cfg      config.DB
	useIndex bool
	mu       sync.RWMutex
}

func New(cfg config.DB, useIndex bool) (*JSONDatabase, error) {
	f, err := os.OpenFile(cfg.DBPath, os.O_CREATE|os.O_RDONLY, 0o666) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("open file error: %w", err)
	}
	defer func(f *os.File) { f.Close() }(f)

	db := make(JSONStorage, EstimatedDBSize)

	dec := json.NewDecoder(f)
	if err := dec.Decode(&db); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("open file error: %w", err)
	}

	index := make(JSONIndex, EstimatedIndexSize)

	if useIndex {
		f, err = os.OpenFile(cfg.IndexPath, os.O_CREATE|os.O_RDONLY, 0o666) //nolint:gomnd
		if err != nil {
			return nil, fmt.Errorf("open file error: %w", err)
		}
		defer f.Close()

		dec = json.NewDecoder(f)
		if err := dec.Decode(&index); err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("open file error: %w", err)
		}
	}

	return &JSONDatabase{
		db:       db,
		index:    index,
		cfg:      cfg,
		mu:       sync.RWMutex{},
		useIndex: useIndex,
	}, nil
}

func (jdb *JSONDatabase) Flush(updateIndex bool) error {
	f, err := os.OpenFile(jdb.cfg.DBPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer func(f *os.File) { f.Close() }(f)

	enc := json.NewEncoder(f)

	if err := enc.Encode(jdb.db); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	if updateIndex {
		f, err = os.OpenFile(jdb.cfg.IndexPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666) //nolint:gomnd
		if err != nil {
			return fmt.Errorf("open file error: %w", err)
		}
		defer f.Close()

		jdb.updateIndex()

		enc = json.NewEncoder(f)

		if err := enc.Encode(jdb.index); err != nil {
			return fmt.Errorf("encode error: %w", err)
		}
	}

	return nil
}

func (jdb *JSONDatabase) AddOne(ci database.ComicsInfo) {
	jdb.mu.Lock()
	jdb.db[ci.ID] = ci
	jdb.mu.Unlock()
}

func (jdb *JSONDatabase) GetByID(id string) (database.ComicsInfo, error) {
	jdb.mu.RLock()
	defer jdb.mu.RUnlock()

	ci, ok := jdb.db[id]
	if !ok {
		return database.ComicsInfo{}, database.ErrNotFound
	}

	return ci, nil
}

func (jdb *JSONDatabase) GetByWord(word string) []string {
	switch jdb.useIndex {
	case true:
		return jdb.index[word]
	default:
		result := make([]string, 0, app.ResultLen)

		for k, c := range jdb.db {
			if slices.Contains(c.Keywords, word) {
				result = append(result, k)
			}
		}

		return result
	}
}

func (jdb *JSONDatabase) updateIndex() {
	for _, c := range jdb.db {
		for _, w := range c.Keywords {
			jdb.index[w] = append(jdb.index[w], c.ID)
		}
	}
}
