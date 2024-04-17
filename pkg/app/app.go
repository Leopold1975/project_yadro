package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

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
	done := make(chan struct{})
	ids := make(chan string, a.cfg.Parallel)
	comicsModels := make(chan xkcd.Model, a.cfg.Parallel)

	go func() {
		a.FetchIDs(ids, done)
	}()

	go func() {
		a.GetComics(ctx, comicsModels, ids, done)
	}()

	go func() {
		a.SaveComics(ctx, comicsModels)
	}()

	select {
	case <-ctx.Done():
		select {
		case <-done:
		default:
			close(done)
		}

		a.Shutdown()

		if a.client.Err != nil {
			return fmt.Errorf("client error: %w", a.client.Err)
		}

		return fmt.Errorf("context error: %w", ctx.Err())
	case <-done:
		a.db.Flush()

		return nil
	}
}

func (a App) Shutdown() {
	log.Println("shutdown app...")
	a.db.Flush()
}

func (a App) FetchIDs(ids chan string, done chan struct{}) {
	defer close(ids)

	for i := 1; ; i++ {
		select {
		case <-done:
			return
		default:
			id := strconv.Itoa(i)
			_, err := a.db.GetByID(id)

			if err != nil && errors.Is(err, database.ErrNotFound) {
				ids <- id
			}
		}
	}
}

func (a App) GetComics(ctx context.Context, comicsModels chan xkcd.Model, ids chan string, done chan struct{}) {
	errCh := make(chan error, ErrorCapacity)
	notFoundErr := make(chan error, a.cfg.Parallel*3) //nolint:gomnd

	defer close(comicsModels)
	defer close(notFoundErr)

	wg := sync.WaitGroup{}
	wg.Add(a.cfg.Parallel + 1)

	go func() {
		defer wg.Done()
		a.client.HandleErrorChan(ctx, errCh)
	}()

	for i := 0; i < a.cfg.Parallel; i++ {
		go func() {
			defer wg.Done()

			for id := range ids {
				ctx, cancel := context.WithTimeout(ctx, time.Second*5) //nolint:gomnd

				comicsModel, err := a.client.GetComics(ctx, id)

				cancel()

				if err != nil {
					if errors.Is(err, xkcd.ErrNotFound) {
						select {
						case notFoundErr <- err:
						case <-done:
							return
						default:
							close(done)
						}

						continue
					}
					errCh <- fmt.Errorf("id: %s error: %w", id, err)

					continue
				}
				select {
				case <-notFoundErr:
				case <-done:
					return
				default:
				}

				comicsModels <- comicsModel
			}
		}()
	}
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
