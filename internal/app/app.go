package app

import (
	"context"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: remove on finixhing dev.
	"os"
	"time"

	"github.com/Leopold1975/yadro_app/internal/controller/httpserver"
	"github.com/Leopold1975/yadro_app/internal/controller/httpserver/middlewares"
	"github.com/Leopold1975/yadro_app/internal/database/jsondb"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
	"github.com/Leopold1975/yadro_app/internal/usecase"
	"github.com/Leopold1975/yadro_app/pkg/logger"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

func Run(ctx context.Context, cfg config.Config, useIndex bool) {
	go func() {
		http.ListenAndServe("localhost:6060", nil) //nolint:gosec,errcheck // pprof
	}()

	lg := logger.New(cfg.Log)

	db, err := jsondb.New(cfg.DB, useIndex)
	if err != nil {
		lg.Error("json db error", err)
		os.Exit(1)
	}

	c := xkcd.New(cfg.SourceURL)

	fetch := usecase.NewComicsFetch(c, db, cfg)

	refresh := usecase.NewBackgroundRefresh(fetch, cfg.RefreshInterval)

	go refresh.Refresh(ctx, lg)

	find := usecase.NewComicsFind(db)

	router := httpserver.NewRouter(find, fetch)

	logRouter := middlewares.LogMiddleware(router, lg)

	serv := httpserver.New(cfg.Server, logRouter)

	go func() {
		if err := serv.Start(); err != nil {
			lg.Error("server start error", err)
		}
	}()

	lg.Info("Server started", "addr", cfg.Server.Addr)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) //nolint:gomnd
	defer cancel()

	if err := serv.Stop(ctx); err != nil { //nolint:contextcheck
		lg.Error("server stop error", err)
	}
}
