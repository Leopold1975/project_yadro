package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Leopold1975/yadro_app/pkg/config"
	"github.com/Leopold1975/yadro_app/pkg/database/jsondb"
	"github.com/Leopold1975/yadro_app/pkg/xkcd"
)

func main() {
	var outInStdout bool

	var numShown int

	flag.BoolVar(&outInStdout, "o", false, "shows saved comics's information, can be used with -n")
	flag.IntVar(&numShown, "n", -1, "show n saved comics' info")

	flag.Parse()

	cfg, err := config.New("config.yaml")
	if err != nil {
		log.Fatalf("config getting error: %s", err.Error())
	}

	db, err := jsondb.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("json db error: %s", err.Error())
	}

	if outInStdout {
		if err := ShowMode(numShown, db); err != nil {
			log.Fatalf("show comics' info error: %s", err.Error())
		}
	} else {
		if err := StoreMode(cfg.SourceURL, db); err != nil {
			log.Fatalf("store comics error: %s", err.Error())
		}
	}
}

func ShowMode(numShown int, db *jsondb.JSONDatabase) error {
	enc := json.NewEncoder(os.Stdout)

	enc.SetIndent("", "\t")

	res := db.GetN(numShown)

	if err := enc.Encode(res); err != nil {
		return fmt.Errorf("encode error: %w", err)
	}

	return nil
}

func StoreMode(sourceURL string, db *jsondb.JSONDatabase) error {
	c := xkcd.New(sourceURL)

	l, err := c.GetAllComics()
	if err != nil {
		return err //nolint:wrapcheck
	}

	infos, err := xkcd.ToDBComicsInfos(l)
	if err != nil {
		log.Printf("xkcd error: %s", err.Error())
	}

	if err := db.CreateList(infos); err != nil {
		return fmt.Errorf("store error: %w", err)
	}

	return nil
}
