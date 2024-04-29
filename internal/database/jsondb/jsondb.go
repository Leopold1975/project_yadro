package jsondb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/Leopold1975/yadro_app/internal/models"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
)

const (
	EstimatedDBSize    = 3000
	EstimatedIndexSize = 20000
)

// we use map here because it's only json db's inner
// data object.
// External clients should use only lists (slices) and models.
type (
	JSONStorage map[string]models.ComicsInfo
	JSONIndex   map[string][]string
)

type JSONmodels struct {
	db       JSONStorage
	index    JSONIndex
	cfg      config.DB
	useIndex bool
	mu       sync.RWMutex
	total    int
	new      int
}

func New(cfg config.DB, useIndex bool) (*JSONmodels, error) {
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

	return &JSONmodels{
		db:       db,
		index:    index,
		cfg:      cfg,
		mu:       sync.RWMutex{},
		useIndex: useIndex,
		total:    len(db),
		new:      0,
	}, nil
}

func (jdb *JSONmodels) Flush(updateIndex bool) (int, int, error) {
	jdb.total = len(jdb.db)

	f, err := os.OpenFile(jdb.cfg.DBPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666) //nolint:gomnd
	if err != nil {
		return 0, 0, fmt.Errorf("open file error: %w", err)
	}
	defer func(f *os.File) { f.Close() }(f)

	enc := json.NewEncoder(f)

	if err := enc.Encode(jdb.db); err != nil {
		return 0, 0, fmt.Errorf("encode error: %w", err)
	}

	if updateIndex {
		f, err = os.OpenFile(jdb.cfg.IndexPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666) //nolint:gomnd
		if err != nil {
			return 0, 0, fmt.Errorf("open file error: %w", err)
		}
		defer f.Close()

		jdb.updateIndex()

		enc = json.NewEncoder(f)

		if err := enc.Encode(jdb.index); err != nil {
			return 0, 0, fmt.Errorf("encode error: %w", err)
		}
	}

	return jdb.total, jdb.new, nil
}

func (jdb *JSONmodels) AddOne(ci models.ComicsInfo) {
	jdb.mu.Lock()

	jdb.db[ci.ID] = ci
	jdb.new++

	jdb.mu.Unlock()
}

func (jdb *JSONmodels) GetByID(id string) (models.ComicsInfo, error) {
	jdb.mu.RLock()
	defer jdb.mu.RUnlock()

	ci, ok := jdb.db[id]
	if !ok {
		return models.ComicsInfo{}, models.ErrNotFound
	}

	return ci, nil
}

func (jdb *JSONmodels) GetByWord(word string, resultLen int) []models.ComicsInfo {
	result := make([]models.ComicsInfo, 0, resultLen)

	switch jdb.useIndex {
	case true:
		ids := jdb.index[word]

		for _, cID := range ids {
			c, err := jdb.GetByID(cID)
			if err != nil {
				continue
			}

			result = append(result, c)
		}

		return result
	default:
		for _, c := range jdb.db {
			if slices.Contains(c.Keywords, word) {
				result = append(result, c)
			}
		}

		return result
	}
}

func (jdb *JSONmodels) updateIndex() {
	for _, c := range jdb.db {
		for _, w := range c.Keywords {
			jdb.index[w] = append(jdb.index[w], c.ID)
		}
	}
}
