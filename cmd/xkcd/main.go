package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Leopold1975/yadro_app/pkg/app"
	"github.com/Leopold1975/yadro_app/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/database/jsondb"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

func main() {
	var configPath string

	flag.StringVar(&configPath, "c", "", "path to configuration file")
	flag.Parse()

	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("config getting error: %s", err.Error())
	}

	db, err := jsondb.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("json db error: %s", err.Error())
	}

	c := xkcd.New(cfg.SourceURL, cfg.Parallel)

	app := app.New(c, db, cfg)

	shutdownSignals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT}

	ctx, cancel := signal.NotifyContext(context.Background(), shutdownSignals...)
	defer cancel()

	err = app.Run(ctx)
	if err != nil {
		log.Println(err)
	}
}
