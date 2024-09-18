package log

import (
	"context"
	"log/slog"
	"os"
	"time"
)

var slogLevel = new(slog.LevelVar) // Info by default

const (
	timeFormat = "2006-01-02 03:04:05PM"

	LogLevelVerbose slog.Level = -6
	LogLevelDebug   slog.Level = -4
	LogLevelInfo    slog.Level = 0
	LogLevelWarn    slog.Level = 4
	LogLevelError   slog.Level = 8
	LogLevelFatal   slog.Level = 12
)

// NewSlogger() creates a new slog instance. module is intended to be the "service" of which the
// logger is apart of, ie "api" or "webui". module can be empty.
func NewSlogger(ctx context.Context, logLevel slog.Level, module string) *slog.Logger {
	ReplaceAttr := func(_ []string, a slog.Attr) slog.Attr {
		// include source in debug level
		if a.Key == "source" && ((slogLevel.Level() != slog.LevelDebug) || (slogLevel.Level() != slog.LevelError)) {
			return slog.Attr{}
		}

		// replace time with more friendly, localized timestamp
		if a.Key == slog.TimeKey {
			return slog.Attr{Key: a.Key, Value: slog.StringValue(time.Now().Format(timeFormat))}
		}

		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			switch {
			case level <= LogLevelVerbose:
				a.Value = slog.StringValue("VERBOSE")
			case level == LogLevelDebug:
				a.Value = slog.StringValue("DEBUG")
			case level == LogLevelInfo:
				a.Value = slog.StringValue("INFO")
			case level == LogLevelWarn:
				a.Value = slog.StringValue("WARN")
			case level == LogLevelError:
				a.Value = slog.StringValue("ERROR")
			case level >= LogLevelFatal:
				a.Value = slog.StringValue("FATAL")
			default:
				a.Value = slog.StringValue("EMERGENCY")
			}

			return a
		}

		return a
	}

	opts := &slog.HandlerOptions{
		Level:       logLevel,
		ReplaceAttr: ReplaceAttr,
	}

	if module == "" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts)).With("module", module)
}
