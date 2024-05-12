package pgtools

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	errCh := make(chan error)
	db := new(pgxpool.Pool)

	go func() {
		defer close(errCh)

		dbc, err := pgxpool.New(ctx, connString)
		if err != nil {
			errCh <- fmt.Errorf("cannot create db pool error: %w", err)

			return
		}

		defaultDelay := time.Second

		for {
			if err := dbc.Ping(ctx); err != nil {
				time.Sleep(defaultDelay)
				defaultDelay += time.Second

				if defaultDelay > time.Second*10 {
					errCh <- fmt.Errorf("cannot ping db error: %w", err)

					return
				}

				continue
			}

			break
		}

		db = dbc
	}()
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context error: %w", ctx.Err())
	case err := <-errCh:
		if err != nil {
			return nil, err
		}

		return db, nil
	}
}
