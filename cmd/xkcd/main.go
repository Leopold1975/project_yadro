package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Leopold1975/yadro_app/internal/app"
	"github.com/Leopold1975/yadro_app/internal/pkg/config"
)

func main() {
	var configPath string

	var phraseString string

	var useIndex bool

	flag.StringVar(&configPath, "c", "", "path to configuration file")
	flag.StringVar(&phraseString, "s", "", "words to find IDs for")
	flag.BoolVar(&useIndex, "i", false, "make db search through index")
	flag.Parse()

	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatalf("config getting error: %s", err.Error())
	}

	shutdownSignals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT}

	ctx, cancel := signal.NotifyContext(context.Background(), shutdownSignals...)
	defer cancel()

	app.Run(ctx, cfg, useIndex)
}
