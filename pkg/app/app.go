package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: Убрать при деплое.
	"strconv"
	"sync"

	"github.com/Leopold1975/yadro_app/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/database"
	"github.com/Leopold1975/yadro_app/pkg/database/jsondb"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

type App struct {
	cfg    config.Config
	client *xkcd.Client
	db     *jsondb.JSONDatabase
}

func New(client *xkcd.Client, db *jsondb.JSONDatabase, cfg config.Config) App {
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

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		a.Shutdown()

		if a.client.Err != nil {
			return fmt.Errorf("client error: %w", a.client.Err)
		}

		return fmt.Errorf("context error: %w", ctx.Err())
	case <-done:
		if err := a.db.Flush(); err != nil {
			return fmt.Errorf("flush to file error %w", err)
		}

		if a.client.Err != nil {
			return fmt.Errorf("client error: %w", a.client.Err)
		}

		return nil
	}
}

func (a App) Shutdown() {
	log.Println("shutdown app...")
	a.db.Flush()
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

	go func() {
		defer wg.Done()
		a.client.HandleErrorChan(ctx, errCh)
	}()

	for i := 0; i < a.cfg.Parallel; i++ {
		go func() {
			defer wg.Done()

			for id := range ids {
				comicsModel, err := a.client.GetComics(ctx, id)
				if err != nil {
					if errors.Is(err, xkcd.ErrNotFound) {
						select {
						case notFoundErr <- err:
						default:
							return
						}
					}
					errCh <- fmt.Errorf("id: %s error: %w", id, err)

					continue
				}
				select {
				case <-notFoundErr: // Возможно получение "ложных" NotFound (из-за удаления id),
				// которые не должны влиять на получение остальных комиксов.
				case <-ctx.Done():
					return
				default:
				}
				comicsModels <- comicsModel
			}
		}()
	}
	wg.Wait()
	close(errCh)
	wg.Add(1) // Добавляем ожидание заверешния функции-обработчика сетевых ошибок.

	wg.Wait()
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
