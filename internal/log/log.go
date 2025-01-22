package log

import (
	"log/slog"
	"os"
	"strings"
)

const (
	LogLevelVerbose slog.Level = -6
	LogLevelDebug   slog.Level = -4
	LogLevelInfo    slog.Level = 0
	LogLevelWarn    slog.Level = 4
	LogLevelError   slog.Level = 8
	LogLevelFatal   slog.Level = 12
)

var levelTypes = map[slog.Leveler]string{
	LogLevelVerbose: "VERBOSE",
	LogLevelDebug:   "DEBUG",
	LogLevelInfo:    "INFO",
	LogLevelWarn:    "WARN",
	LogLevelError:   "ERROR",
	LogLevelFatal:   "FATAL",
}

func New(logLevel slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey {
					level := a.Value.Any().(slog.Level)
					levelLabel, exists := levelTypes[level]
					if !exists {
						levelLabel = level.String()
					}

					a.Value = slog.StringValue(levelLabel)
				}

				return a
			},
		}))
}

// Valid values are "info", "error", "debug" and "fatal".
// Returns "info" if given invalid log level.
func StringToLogLevel(logType string) slog.Level {
	switch {
	default:
		return LogLevelInfo
	case strings.EqualFold(logType, "info"):
		return LogLevelInfo
	case strings.EqualFold(logType, "warn"):
		return LogLevelWarn
	case strings.EqualFold(logType, "verbose"):
		return LogLevelVerbose
	case strings.EqualFold(logType, "debug"):
		return LogLevelDebug
	case strings.EqualFold(logType, "error"):
		return LogLevelError
	case strings.EqualFold(logType, "fatal"):
		return LogLevelFatal
	}
}
