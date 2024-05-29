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
			defaultDelay := time.Second

			for err != nil {
				time.Sleep(defaultDelay)
				defaultDelay += time.Second

				if defaultDelay > time.Second*10 {
					errCh <- fmt.Errorf("cannot connect to db error: %w", err)

					return
				}

				dbc, err = pgxpool.New(ctx, connString)
			}
		}

		defaultDelay := time.Second

		if err := dbc.Ping(ctx); err != nil {
			for err != nil {
				time.Sleep(defaultDelay)
				defaultDelay += time.Second

				if defaultDelay > time.Second*10 {
					errCh <- fmt.Errorf("cannot ping db error: %w", err)

					return
				}

				err = dbc.Ping(ctx)
			}
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
