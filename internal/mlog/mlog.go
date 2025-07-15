package mlog

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

var (
	logger *slog.Logger
)

func onlyMessage(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.MessageKey && len(groups) == 0 {
		return a
	}
	return slog.Attr{}
}

func Init(level slog.Level) {
	logger = slog.New(
		tint.NewHandler(
			os.Stderr,
			&tint.Options{
				ReplaceAttr: onlyMessage,
				Level:       level,
			},
		),
	)
	slog.SetDefault(logger)

}

func GetLogger() *slog.Logger {
	return logger
}
