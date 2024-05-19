package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Leopold1975/yadro_app/internal/auth/models"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/pgtools"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DB) (UserRepo, error) {
	connString := "postgres://" + cfg.Username + ":" + cfg.Password + "@" +
		cfg.Addr + "/" + cfg.DB + "?" + "sslmode=" + cfg.SSLmode + "&pool_max_conns=" + cfg.MaxConns

	db, err := pgtools.Connect(ctx, connString)
	if err != nil {
		return UserRepo{}, fmt.Errorf("connect to db error: %w", err)
	}

	connString = "postgres://" + cfg.Username + ":" + cfg.Password + "@" +
		cfg.Addr + "/" + cfg.DB + "?sslmode=" + cfg.SSLmode

	if err := pgtools.ApplyMigration(connString, cfg.Version); err != nil {
		return UserRepo{}, fmt.Errorf("apply migration error: %w", err)
	}

	return UserRepo{
		db: db,
	}, nil
}

func (ur *UserRepo) GetUser(ctx context.Context, username string) (models.User, error) {
	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pb.Select("id", "username", "passwordHash", "role").
		From("users").
		Where(squirrel.Eq{"username": username}).ToSql()
	if err != nil {
		return models.User{}, fmt.Errorf("to sql error %w", err)
	}

	row := ur.db.QueryRow(ctx, query, args...)

	var u models.User

	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
	}

	return u, nil
}

func (ur *UserRepo) CreateUser(ctx context.Context, user models.User) error {
	tx, err := ur.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx error %w", err)
	}

	defer func() {
		err = pgtools.CommitOrRollback(ctx, tx, err, "create user")
	}()

	pb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := pb.Insert("users").Columns("username", "passwordHash", "role").
		Values(user.Username, user.PasswordHash, user.Role).ToSql()
	if err != nil {
		return fmt.Errorf("to sql error %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("exec error %w", err)
	}

	return nil
}
