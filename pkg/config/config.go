package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SourceURL string `yaml:"source_url" env-default:"https://xkcd.com"` //nolint:tagliatelle
	DBPath    string `yaml:"db_file" env-default:"database.json"`       //nolint:tagliatelle
}

func New(path string) (Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot read config error: %w", err)
	}

	return cfg, nil
}
