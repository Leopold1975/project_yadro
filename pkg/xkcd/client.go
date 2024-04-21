package xkcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	infoSuffix = "info.0.json"
)

type Client struct {
	sourceURL string
	client    *http.Client
	Err       error
}

func New(sourceURL string) *Client {
	return &Client{
		sourceURL: sourceURL,
		client:    http.DefaultClient,
		Err:       nil,
	}
}

func (c *Client) GetComics(ctx context.Context, id string) (Model, error) {
	resURL, err := url.JoinPath(c.sourceURL, id, infoSuffix)
	if err != nil {
		return Model{}, fmt.Errorf("join path error: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5) //nolint:gomnd
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

func (c *Client) HandleErrorChan(ctx context.Context, errCh <-chan error) {
	var err error

loop:
	for e := range errCh {
		select {
		case <-ctx.Done():
			break loop
		default:
			if e != nil {
				err = errors.Join(err, e)
			}
		}
	}

	if err != nil {
		c.Err = err
	}
}
