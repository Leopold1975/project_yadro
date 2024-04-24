package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: remove on finixhing dev.
	"sort"
	"strconv"
	"sync"

	"github.com/Leopold1975/yadro_app/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/database"
	"github.com/Leopold1975/yadro_app/pkg/words"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

type App struct {
	cfg    config.Config
	client *xkcd.Client
	db     Storage
}

const (
	updateIndex = true
	ResultLen   = 10
)

type Storage interface {
	AddOne(ci database.ComicsInfo)
	GetByID(id string) (database.ComicsInfo, error)
	GetByWord(word string) []string
	Flush(updateIndex bool) error
}

func New(client *xkcd.Client, db Storage, cfg config.Config) App {
	return App{
		client: client,
		db:     db,
		cfg:    cfg,
	}
}

const (
	ErrorCapacity   = 10
	NotFoundAllowed = 20
)

func (a App) Run(ctx context.Context) error {
	ids := make(chan string, a.cfg.Parallel)
	comicsModels := make(chan xkcd.Model, a.cfg.Parallel)

	go func() {
		http.ListenAndServe("localhost:6060", nil) //nolint:gosec,errcheck // pprof
	}()

	wg := sync.WaitGroup{}
	wg.Add(3) //nolint:gomnd

	ctxF, cancel := context.WithCancel(ctx)
	go func() {
		defer wg.Done()
		a.FetchIDs(ctxF, ids)
	}()

	go func() {
		defer wg.Done()
		a.GetComics(ctx, comicsModels, ids)
		cancel()
	}()

	go func() {
		defer wg.Done()
		a.SaveComics(ctx, comicsModels)
	}()

	wg.Wait()

	if err := a.db.Flush(updateIndex); err != nil {
		return fmt.Errorf("flush to file error %w", err)
	}

	if a.client.Err != nil {
		return fmt.Errorf("client error: %w", a.client.Err)
	}

	return nil
}

func (a App) FetchIDs(ctx context.Context, ids chan string) {
	defer close(ids)

	for i := 1; ; i++ {
		id := strconv.Itoa(i)

		_, err := a.db.GetByID(id)
		if err != nil && errors.Is(err, database.ErrNotFound) {
			select {
			case <-ctx.Done():
				return
			case ids <- id:
			}
		}
	}
}

func (a App) GetComics(ctx context.Context, comicsModels chan<- xkcd.Model, ids <-chan string) {
	errCh := make(chan error, ErrorCapacity)
	notFoundErr := make(chan error, a.cfg.Parallel*3) //nolint:gomnd
	// Канал для сбора ошибок NotFOund, переполнение которого считается сигналом к окончанию работы.

	defer close(comicsModels)
	defer close(notFoundErr)

	wg := sync.WaitGroup{}
	wg.Add(a.cfg.Parallel)

	doneErr := make(chan struct{})
	go func() {
		a.client.HandleErrorChan(ctx, errCh)
		close(doneErr)
	}()

	for i := 0; i < a.cfg.Parallel; i++ {
		go func() {
			defer wg.Done()
			a.getComicsParallel(ctx, comicsModels, ids, notFoundErr, errCh)
		}()
	}
	wg.Wait()

	close(errCh)
	<-doneErr
}

func (a App) SaveComics(ctx context.Context, comicsModels chan xkcd.Model) {
	for cm := range comicsModels {
		select {
		case <-ctx.Done():
			return
		default:
			ci, err := xkcd.ToDBComicsInfo(cm)
			if err != nil {
				log.Println("ToDBComicsInfo error: ", err)
			}

			a.db.AddOne(ci)
		}
	}
}

func (a App) GetIDs(phrase string) ([]string, error) {
	normalizedPhrase, err := words.StemWords(phrase)
	if err != nil {
		return nil, fmt.Errorf("stem words error: %w", err)
	}

	ids := make([][]string, 0, len(normalizedPhrase))
	for _, word := range normalizedPhrase {
		ids = append(ids, a.db.GetByWord(word))
	}

	result := GetTopIDs(ids...)

	return result, nil
}

func (a App) Health() error {
	var err error

	for i := 1; i < 2922; i++ {
		if i == 404 { //nolint:gomnd
			continue
		}

		_, e := a.db.GetByID(strconv.Itoa(i))
		if e != nil {
			err = errors.Join(err, fmt.Errorf("%w id %d", e, i))
		}
	}

	return err
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

func (a App) getComicsParallel(ctx context.Context, //nolint:cyclop
	comicsModels chan<- xkcd.Model, ids <-chan string, notFoundErr chan error, errCh chan error,
) {
	for id := range ids {
		comicsModel, err := a.client.GetComics(ctx, id)
		if err != nil {
			if errors.Is(err, xkcd.ErrNotFound) {
				select {
				case notFoundErr <- err:
				default:
					return
				}

				continue
			}
			select {
			case errCh <- fmt.Errorf("id: %s error: %w", id, err):
			default:
			}

			continue
		}

		select {
		case <-notFoundErr: // Возможно получение "ложных" NotFound (из-за удаления id),
		// которые не должны влиять на получение остальных комиксов.
		default:
		}

		select {
		case <-ctx.Done():
			return
		case comicsModels <- comicsModel:
		}
	}
}
