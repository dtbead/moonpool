package log

import (
	"context"
	"log/slog"
	"os"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

var slogLevel = new(slog.LevelVar) // Info by default

func NewSlogLogger(ctx context.Context) Logger {
	ReplaceAttr := func(_ []string, a slog.Attr) slog.Attr {
		// Hack-y way of omitting AddSource option from
		// logging unless it's set to debug.
		if a.Key == "source" && ((slogLevel.Level() != slog.LevelDebug) || (slogLevel.Level() != slog.LevelError)) {
			return slog.Attr{}
		}
		return slog.Attr{Key: a.Key, Value: a.Value}
	}

	opts := &slog.HandlerOptions{
		Level:       slogLevel,
		ReplaceAttr: ReplaceAttr,
	}

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}
