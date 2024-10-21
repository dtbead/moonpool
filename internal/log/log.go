package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
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
// logger is a part of to better keep track of logging components, ie "api" or "webui".
// module can be safely set to empty.
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
			case level == LogLevelError:
				a.Value = slog.StringValue("ERROR")
			default:
				a.Value = slog.StringValue("UNKNOWN")
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

// Valid values are "info", "error", "debug", "debug" and "fatal".
// Returns "info" if given invalid log level.
func StringToLogLevel(logType string) slog.Level {
	switch {
	default:
		return LogLevelInfo
	case strings.EqualFold(logType, "info"):
		return LogLevelInfo
	case strings.EqualFold(logType, "verbose"):
		return LogLevelVerbose
	case strings.EqualFold(logType, "error"):
		return LogLevelError
	case strings.EqualFold(logType, "fatal"):
		return LogLevelFatal
	}
}
