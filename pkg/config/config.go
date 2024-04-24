package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SourceURL string `env-default:"https://xkcd.com" yaml:"source_url"` //nolint:tagliatelle
	Parallel  int    `env-required:"true"            yaml:"parallel"`
	DB        DB     `yaml:"db"`
}

type DB struct {
	DBPath    string `env-default:"database.json" yaml:"db_file"` //nolint:tagliatelle
	IndexPath string `yaml:"indexPath"`
}

func New(path string) (Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot read config error: %w", err)
	}

	return cfg, nil
}
