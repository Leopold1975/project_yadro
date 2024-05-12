package pgtools

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func CommitOrRollback(ctx context.Context, tx pgx.Tx, err error, where string) error {
	if err == nil {
		if errT := tx.Commit(ctx); errT != nil {
			err = fmt.Errorf("commit error: %w", errT)
		}
	} else {
		if errT := tx.Rollback(ctx); errT != nil {
			err = fmt.Errorf("%s error: %w rollback error: %w", where, err, errT)
		} else {
			err = fmt.Errorf("%s error: %w", where, err)
		}
	}

	return err
}
