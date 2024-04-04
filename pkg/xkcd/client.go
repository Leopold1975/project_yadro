package xkcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	infoSuffix = "info.0.json"
)

const (
	NumWorkers      = 100
	ErrorCapacity   = 10
	NotFoundAllowed = 20
)

type Client struct {
	sourceURL string
	client    *http.Client
}

func New(sourceURL string) *Client {
	return &Client{
		sourceURL: sourceURL,
		client:    http.DefaultClient,
	}
}

func (c *Client) GetComics(id string) (Model, error) {
	resURL, err := url.JoinPath(c.sourceURL, id, infoSuffix)
	if err != nil {
		return Model{}, fmt.Errorf("join path error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) //nolint:gomnd
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resURL, nil)
	if err != nil {
		return Model{}, fmt.Errorf("HTTP GET error: %w", err)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return Model{}, fmt.Errorf("HTTP GET error: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return Model{}, ErrNotFound
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return Model{}, fmt.Errorf("read body error: %w", err)
		}

		var m Model
		if err := json.Unmarshal(body, &m); err != nil {
			return Model{}, fmt.Errorf("unmarshal body error: %w", err)
		}

		return m, nil
	default:
		return Model{}, fmt.Errorf("id: %s code: %d err: %w", id, resp.StatusCode, ErrUnexpectedCode)
	}
}

func (c *Client) GetAllComics() ([]Model, error) {
	result := make([]Model, 0, 2048) //nolint:gomnd

	done := make(chan struct{})
	errCh := make(chan error, ErrorCapacity)
	notFoundErr := make(chan error, NotFoundAllowed)

	comics := make(chan Model, NumWorkers)
	ids := make(chan string, NumWorkers*3) //nolint:gomnd

	go func() {
		fillIDs(ids, done)
	}()

	go func() {
		defer close(comics)
		defer close(errCh)
		defer close(notFoundErr)

		wg := sync.WaitGroup{}
		wg.Add(NumWorkers)

		for i := 0; i < NumWorkers; i++ {
			go func() {
				defer wg.Done()

				for id := range ids {
					m, err := c.GetComics(id)
					if err != nil {
						if errors.Is(err, ErrNotFound) {
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
					comics <- m
				}
			}()
		}
		wg.Wait()
	}()

	var err error

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		err = handleErrorChan(errCh, &wg)
	}()

	for c := range comics {
		result = append(result, c)
	}

	wg.Wait()

	if err != nil {
		return []Model{}, err
	}

	return result, nil
}

func handleErrorChan(errCh <-chan error, wg *sync.WaitGroup) error {
	defer wg.Done()

	var err error

	for e := range errCh {
		if e != nil {
			err = errors.Join(err, e)
		}
	}

	return err
}

func fillIDs(ids chan<- string, done <-chan struct{}) {
	for i := 1; ; i++ {
		select {
		case <-done:
			close(ids)

			return
		default:
			select {
			case ids <- strconv.Itoa(i):
			case <-done:
				close(ids)

				return
			}
		}
	}
}
