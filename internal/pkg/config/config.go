package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SourceURL      string         `env-default:"https://xkcd.com" yaml:"source_url"` //nolint:tagliatelle
	Parallel       Parallel       `env-required:"true"            yaml:"parallel"`
	APIConcurrency APIConcurrency `env-required:"true"            yaml:"apiConcurrency"`
	DB             DB             `yaml:"db"`
	Log            LogLvl         `yaml:"log"`
	Server         Server         `yaml:"server"`
	RefreshTime    RefreshTime    `yaml:"refreshTime"`
}

type DB struct {
	Addr     string `yaml:"addr"`
	Username string `env:"POSTGRES_USER"     env-required:"true" yaml:"username"`
	Password string `env:"POSTGRES_PASSWORD" yaml:"password"`
	DB       string `env:"POSTGRES_DB"       env-required:"true" yaml:"db"`
	SSLmode  string `yaml:"sslmode"`
	MaxConns string `yaml:"maxConns"`
	Version  int    `yaml:"version"`
}

type Server struct {
	Addr         string        `yaml:"addr"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

type LogLvl string

type Parallel int

type APIConcurrency int

func New(path string) (Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot read config error: %w", err)
	}

	return cfg, nil
}

type RefreshTime struct {
	time.Time
}

func (rt *RefreshTime) UnmarshalText(text []byte) error {
	t, err := time.Parse("15:04:05 Z0700", string(text))
	if err != nil {
		return fmt.Errorf("time parse error %w", err)
	}

	rt.Time = t

	return nil
}
