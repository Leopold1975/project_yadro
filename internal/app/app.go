package app

import (
	"context"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: remove on finixhing dev.
	"os"
	"time"

	"github.com/Leopold1975/yadro_app/internal/controller/httpserver"
	"github.com/Leopold1975/yadro_app/internal/controller/httpserver/middlewares"
	"github.com/Leopold1975/yadro_app/internal/database/postgresdb"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/internal/usecase"
	"github.com/Leopold1975/yadro_app/pkg/logger"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

const (
	JSONDB     = "json"
	PostgresDB = "postgres"
)

func Run(ctx context.Context, cfg config.Config, useIndex bool) {
	go func() {
		http.ListenAndServe("localhost:6060", nil) //nolint:gosec,errcheck // pprof
	}()

	lg := logger.New(cfg.Log)

	db, err := postgresdb.New(ctx, cfg.DB, useIndex)
	if err != nil {
		lg.Error("postgres db error", "error", err)
		os.Exit(1)
	}

	c := xkcd.New(cfg.SourceURL)

	fetch := usecase.NewComicsFetch(c, &db, cfg.Parallel, lg)

	refresh := usecase.NewBackgroundRefresh(fetch, cfg.RefreshTime.Time)

	go refresh.Refresh(ctx, lg)

	find := usecase.NewComicsFind(&db, lg)

	router := httpserver.NewRouter(find, fetch)

	clmw := middlewares.NewConcurrencyLimitter(cfg.APIConcurrency)
	defer clmw.Close()

	router = middlewares.LogMiddleware(
		clmw.ConcurrencyMiddleware(router, cfg.APIConcurrency),
		lg)

	serv := httpserver.New(cfg.Server, router)

	go func() {
		lg.Info("started server on", "addr", cfg.Server.Addr)

		if err := serv.Start(); err != nil {
			lg.Error("server start error", "error", err)
		}
	}()

	lg.Info("Server started", "addr", cfg.Server.Addr)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) //nolint:gomnd
	defer cancel()

	if err := serv.Stop(ctx); err != nil { //nolint:contextcheck
		lg.Error("server stop error", "error", err)
	}
}
