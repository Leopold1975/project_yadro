package postgresdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/Leopold1975/yadro_app/internal/models"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/pgtools"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // used for driver
)

type ComicsRepo struct {
	db            *pgxpool.Pool
	newComics     atomic.Int32
	useIndexTable bool
}

func New(ctx context.Context, cfg config.DB, useIndexTable bool) (ComicsRepo, error) {
	connString := "postgres://" + cfg.Username + ":" + cfg.Password + "@" +
		cfg.Addr + "/" + cfg.DB + "?" + "sslmode=" + cfg.SSLmode + "&pool_max_conns=" + cfg.MaxConns

	db, err := pgtools.Connect(ctx, connString)
	if err != nil {
		return ComicsRepo{}, fmt.Errorf("connect to db error: %w", err)
	}

	connString = "postgres://" + cfg.Username + ":" + cfg.Password + "@" +
		cfg.Addr + "/" + cfg.DB + "?sslmode=" + cfg.SSLmode

	if err := pgtools.ApplyMigration(connString, cfg.Version); err != nil {
		return ComicsRepo{}, fmt.Errorf("apply migration error: %w", err)
	}

	return ComicsRepo{
		db:            db,
		newComics:     atomic.Int32{},
		useIndexTable: useIndexTable,
	}, nil
}

func (cr *ComicsRepo) AddOne(ctx context.Context, ci models.ComicsInfo) (err error) {
	tx, err := cr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx error %w", err)
	}

	defer func() {
		err = pgtools.CommitOrRollback(ctx, tx, err, "add one")
	}()

	jsonKeywords, err := json.Marshal(ci.Keywords)
	if err != nil {
		return fmt.Errorf("mashal keywords error %w", err)
	}

	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pb.Insert("comics").Columns("id", "url", "keywords").
		Values(ci.ID, ci.URL, string(jsonKeywords)).
		ToSql()
	if err != nil {
		return fmt.Errorf("to sql error %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec error %w", err)
	}

	if err := updateIndex(ctx, tx, ci.Keywords, ci.ID); err != nil {
		return err
	}

	cr.newComics.Add(1)

	return nil
}

func (cr *ComicsRepo) GetByID(ctx context.Context, id string) (models.ComicsInfo, error) {
	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pb.Select("id", "url", "keywords").From("comics").
		Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return models.ComicsInfo{}, fmt.Errorf("to sql error %w", err)
	}

	row := cr.db.QueryRow(ctx, query, args...)

	var ci models.ComicsInfo

	var keywords string

	if err := row.Scan(&ci.ID, &ci.URL, &keywords); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ComicsInfo{}, models.ErrNotFound
		}

		return models.ComicsInfo{}, fmt.Errorf("scan error %w", err)
	}

	if err := json.Unmarshal([]byte(keywords), &ci.Keywords); err != nil {
		return models.ComicsInfo{}, fmt.Errorf("unmarshal error %w", err)
	}

	return ci, nil
}

func (cr *ComicsRepo) GetByWord(ctx context.Context, word string, resultLen int) ([]models.ComicsInfo, error) {
	result := make([]models.ComicsInfo, 0, resultLen)
	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	var query string

	var args []interface{}

	var err error

	if cr.useIndexTable {
		query, args, err = pb.Select("comics.id", "comics.url", "comics.keywords").From("comics").
			Join("keyword_comics_map kc ON kc.comics_id = comics.id").
			Join("keywords k ON k.id = kc.keyword_id").
			Where(squirrel.Eq{"k.keyword": word}).ToSql()
		if err != nil {
			return nil, fmt.Errorf("to sql error %w", err)
		}
	} else {
		query, args, err = pb.Select("id", "url", "keywords").From("comics").
			Where(squirrel.Expr("keywords @> ?", fmt.Sprintf(`"%s"`, word))).ToSql()
		if err != nil {
			return nil, fmt.Errorf("to sql error %w", err)
		}
	}

	rows, err := cr.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query error %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var m models.ComicsInfo

		var keywordsJSON string

		if err := rows.Scan(&m.ID, &m.URL, &keywordsJSON); err != nil {
			return nil, fmt.Errorf("scan error %w", err)
		}

		if err := json.Unmarshal([]byte(keywordsJSON), &m.Keywords); err != nil {
			return nil, fmt.Errorf("unmarshal error %w", err)
		}

		result = append(result, m)
	}

	return result, nil
}

func (cr *ComicsRepo) Flush(ctx context.Context, _ bool) (int, int, error) {
	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, _, err := pb.Select("COUNT(id)").From("comics").ToSql()
	if err != nil {
		return 0, 0, fmt.Errorf("to sql error %w", err)
	}

	row := cr.db.QueryRow(ctx, query)

	var total int

	if err = row.Scan(&total); err != nil {
		return 0, 0, fmt.Errorf("scan error %w", err)
	}

	newC := int(cr.newComics.Load())
	cr.newComics.Store(0)

	return total, newC, nil
}

func updateIndex(ctx context.Context, tx pgx.Tx, keywords []string, comicsID string) error {
	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	for _, keyword := range keywords {
		query, args, err := pb.Insert("keywords").Columns("keyword").
			Values(keyword).
			Suffix("ON CONFLICT (keyword) DO NOTHING"). // Игнорируем конфликт уникальности
			ToSql()
		if err != nil {
			return fmt.Errorf("to sql error %w", err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("exec error %w", err)
		}

		subQuery := squirrel.Select("id").From("keywords").Where(squirrel.Eq{"keyword": keyword})

		query, args, err = pb.Insert("keyword_comics_map").
			Columns("keyword_id", "comics_id").
			Select(subQuery.Column(squirrel.Expr("? AS comics_id", comicsID))).
			Suffix("ON CONFLICT (keyword_id, comics_id) DO NOTHING").
			ToSql()
		if err != nil {
			return fmt.Errorf("to sql error %w", err)
		}

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("exec error %w", err)
		}
	}

	return nil
}
