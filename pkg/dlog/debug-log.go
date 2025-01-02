package dlog

import (
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

type Dlog struct {
	*slog.Logger
}

func Init() Dlog {

	slogOpts := &slog.HandlerOptions{
		AddSource: true,
	}

	slogOpts.Level = slog.LevelDebug

	opts := &devslog.Options{
		HandlerOptions:    slogOpts,
		MaxSlicePrintSize: 4,
		SortKeys:          true,
		NewLineAfterLog:   false,
	}

	logger := slog.New(devslog.NewHandler(os.Stdout, opts))

	slog.SetDefault(logger)

	return Dlog{logger}
}

// debug log
func (d Dlog) Log(msg string, args ...any) {
	d.Debug(msg, args...)
}
