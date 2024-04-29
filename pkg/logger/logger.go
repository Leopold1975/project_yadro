package logger

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/Leopold1975/yadro_app/internal/pkg/config"
)

type Logger struct {
	*slog.Logger
}

const (
	InfoLvl  = "info"
	DebugLvl = "debug"
)

func New(cfg config.LogLvl) Logger {
	var lvl slog.Level

	switch cfg {
	case InfoLvl:
		lvl = slog.LevelInfo
	case DebugLvl:
		lvl = slog.LevelDebug
	default:
		lvl = slog.LevelInfo
	}

	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       lvl,
		ReplaceAttr: Rep,
	}))

	return Logger{l}
}

func Rep(_ []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case "function":
		fns := a.Value.String()

		fn := trimFunc(fns)

		return slog.String(a.Key, fn)
	case "file":
		p := a.Value.String()

		p = trimPath(p)

		return slog.String(a.Key, p)
	default:
		return a
	}
}

func trimFunc(fullFunc string) string {
	funcs := strings.Split(fullFunc, "/")

	if len(funcs) < 3 { //nolint:gomnd
		return fullFunc
	}

	fullFunc = strings.Join(funcs[2:], "/")

	return fullFunc
}

func trimPath(fullPath string) string {
	_, file := path.Split(fullPath)

	return file
}
