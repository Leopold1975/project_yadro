package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/Leopold1975/yadro_app/internal/models"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/logger"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

const (
	ErrorCapacity   = 10
	NotFoundAllowed = 20
	updateIndex     = false
)

type FetchComicsUsecase struct {
	client   *xkcd.Client
	db       Storage
	parallel config.Parallel
	l        logger.Logger
}

type FetchResponse struct {
	New   int
	Total int
}

func NewComicsFetch(client *xkcd.Client, db Storage, parallel config.Parallel, l logger.Logger) FetchComicsUsecase {
	return FetchComicsUsecase{
		client:   client,
		db:       db,
		parallel: parallel,
		l:        l,
	}
}

func (f FetchComicsUsecase) FetchComics(ctx context.Context) (FetchResponse, error) {
	ids := make(chan string, f.parallel)
	comicsModels := make(chan models.XKCDModel, f.parallel)

	wg := sync.WaitGroup{}
	wg.Add(3) //nolint:gomnd

	ctxF, cancel := context.WithCancel(ctx)
	go func() {
		defer wg.Done()
		f.fetchIDs(ctxF, ids)
	}()

	go func() {
		defer wg.Done()
		f.getComics(ctx, comicsModels, ids)
		cancel()
	}()

	go func() {
		defer wg.Done()
		f.saveComics(ctx, comicsModels)
	}()

	wg.Wait()

	totalC, newC, err := f.db.Flush(updateIndex)
	if err != nil {
		return FetchResponse{}, fmt.Errorf("flush to file error %w", err)
	}

	if f.client.Err != nil {
		return FetchResponse{}, fmt.Errorf("client error: %w", f.client.Err)
	}

	return FetchResponse{
		New:   newC,
		Total: totalC,
	}, nil
}

func (f FetchComicsUsecase) fetchIDs(ctx context.Context, ids chan string) {
	defer close(ids)

	for i := 1; ; i++ {
		id := strconv.Itoa(i)

		_, err := f.db.GetByID(id)
		if err != nil && errors.Is(err, models.ErrNotFound) {
			select {
			case <-ctx.Done():
				return
			case ids <- id:
			}
		}
	}
}

func (f FetchComicsUsecase) getComics(ctx context.Context, comicsModels chan<- models.XKCDModel, ids <-chan string) {
	errCh := make(chan error, ErrorCapacity)
	notFoundErr := make(chan error, f.parallel*3) //nolint:gomnd
	// Канал для сбора ошибок NotFOund, переполнение которого считается сигналом к окончанию работы.

	defer close(comicsModels)
	defer close(notFoundErr)

	wg := sync.WaitGroup{}
	wg.Add(int(f.parallel))

	doneErr := make(chan struct{})
	go func() {
		f.client.HandleErrorChan(ctx, errCh)
		close(doneErr)
	}()

	for i := 0; i < int(f.parallel); i++ {
		go func() {
			defer wg.Done()
			f.getComicsParallel(ctx, comicsModels, ids, notFoundErr, errCh)
		}()
	}
	wg.Wait()

	close(errCh)
	<-doneErr
}

func (f FetchComicsUsecase) saveComics(ctx context.Context, comicsModels chan models.XKCDModel) {
	for cm := range comicsModels {
		select {
		case <-ctx.Done():
			return
		default:
			ci, err := models.ToDBComicsInfo(cm)
			if err != nil {
				f.l.Error("ToDBComicsInfo error", "error", err)
			}

			f.db.AddOne(ci)
		}
	}
}

func (f FetchComicsUsecase) getComicsParallel(ctx context.Context, //nolint:cyclop
	comicsModels chan<- models.XKCDModel, ids <-chan string, notFoundErr chan error, errCh chan error,
) {
	for id := range ids {
		comicsModel, err := f.client.GetComics(ctx, id)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
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
