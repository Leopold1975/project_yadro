package app

import (
	"context"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: remove on finixhing dev.
	"time"

	"github.com/Leopold1975/yadro_app/internal/auth/database/postgres"
	auth "github.com/Leopold1975/yadro_app/internal/auth/usecase"
	"github.com/Leopold1975/yadro_app/internal/controller/httpserver"
	"github.com/Leopold1975/yadro_app/internal/controller/httpserver/middlewares"
	"github.com/Leopold1975/yadro_app/internal/database/postgresdb"
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

	db, err := postgresdb.New(ctx, cfg.DB, useIndex)
	if err != nil {
		lg.Error("postgres db error", "error", err)

		return
	}

	userDB, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		lg.Error("postgres db error", "error", err)

		return
	}

	c := xkcd.New(cfg.SourceURL)

	fetch := usecase.NewComicsFetch(c, &db, cfg.Parallel, lg)
	refresh := usecase.NewBackgroundRefresh(fetch, cfg.RefreshTime.Time)
	find := usecase.NewComicsFind(&db, lg)

	go refresh.Refresh(ctx, lg)

	login := auth.NewLoginUser(cfg.Auth, &userDB)
	authUC := auth.NewAuthUser(cfg.Auth, &userDB)

	router := httpserver.NewRouter(find, fetch, login)

	clmw := middlewares.NewConcurrencylimiter(cfg.APIConcurrency)
	defer clmw.Close()

	rl := middlewares.NewRateLimiter(cfg.Ratelimit)

	router = middlewares.LogMiddleware(
		clmw.ConcurrencyMiddleware(
			middlewares.AuthMidleware(
				rl.RatelimiterMiddleware(
					router,
				), authUC),
		),
		lg)

	// router = middlewares.ProfileMiddleware(router)

	serv := httpserver.New(cfg.Server, router)

	go func() {
		lg.Info("started server on", "addr", cfg.Server.Addr)

		if err := serv.Start(); err != nil {
			lg.Error("server start error", "error", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) //nolint:gomnd
	defer cancel()

	if err := serv.Stop(ctx); err != nil { //nolint:contextcheck
		lg.Error("server stop error", "error", err)
	}
}
