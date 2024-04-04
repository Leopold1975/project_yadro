package jsondb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/Leopold1975/yadro_app/pkg/database"
)

const (
	EstimatedSize = 3000
)

// we use map here because it's only json db's inner
// data object.
// External clients should use only lists (slices) and models.
type JSONStorage map[string]database.ComicsInfo

type JSONDatabase struct {
	db       JSONStorage
	filePath string
}

func New(path string) (*JSONDatabase, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0o666) //nolint:gomnd
	if err != nil {
		return nil, fmt.Errorf("open file error: %w", err)
	}
	defer f.Close()

	db := make(JSONStorage, EstimatedSize)

	dec := json.NewDecoder(f)
	if err := dec.Decode(&db); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("open file error: %w", err)
	}

	return &JSONDatabase{
		db:       db,
		filePath: path,
	}, nil
}

func (jdb *JSONDatabase) CreateList(comics []database.ComicsInfo) error {
	f, err := os.OpenFile(jdb.filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o666) //nolint:gomnd
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	defer f.Close()

	db := toJSONStorage(comics)

	enc := json.NewEncoder(f)

	if err := enc.Encode(db); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	return nil
}

func (jdb *JSONDatabase) GetByID(id string) (database.ComicsInfo, error) {
	ci, ok := jdb.db[id]
	if !ok {
		return database.ComicsInfo{}, database.ErrNotFound
	}

	return ci, nil
}

// GetN returns N values with ID <= N.
// N can be set to -1 to return all the values.
func (jdb *JSONDatabase) GetN(n int) JSONStorage {
	if n == -1 {
		return jdb.db
	}

	res := make(JSONStorage, n)

	for i := 1; i <= n; i++ {
		id := strconv.Itoa(i)
		if _, ok := jdb.db[id]; !ok {
			n++

			continue
		}

		res[id] = jdb.db[id]
	}

	return res
}

func toJSONStorage(comics []database.ComicsInfo) JSONStorage {
	m := make(JSONStorage, len(comics))

	for _, c := range comics {
		m[c.ID] = c
	}

	return m
}
