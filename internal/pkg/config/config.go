package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SourceURL       string        `env-default:"https://xkcd.com" yaml:"source_url"` //nolint:tagliatelle
	Parallel        Parallel      `env-required:"true"            yaml:"parallel"`
	DB              DB            `yaml:"db"`
	Log             LogLvl        `yaml:"log"`
	Server          Server        `yaml:"server"`
	RefreshInterval time.Duration `yaml:"refreshInterval"`
}

type DB struct {
	DBPath    string `env-default:"database.json" yaml:"db_file"` //nolint:tagliatelle
	IndexPath string `yaml:"indexPath"`
}

type Server struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

type LogLvl string

type Parallel int

func New(path string) (Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot read config error: %w", err)
	}

	return cfg, nil
}
