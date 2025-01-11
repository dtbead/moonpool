package config

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/dtbead/moonpool/internal/log"
)

const (
	PROFILING_CPU  = "cpu"
	PROFILING_MEM  = "memory"
	PROFILING_NONE = "none"
)

type DynamicWebReloading struct {
	Enable bool
	Path   string
}

type Config struct {
	Debug struct {
		DynamicWebReloading DynamicWebReloading
	}
	Logging struct {
		LogLevel        string
		slogLogLevel    slog.Level
		Profiling       string
		FileLoggingPath string
		FileLogging     bool
	}
	MediaPath     string
	ArchivePath   string
	ThumbnailPath string
	ListenAddress string
	WebUIPort     int
}

// DefaultValues returns a config with sane defaults
func DefaultValues() Config {
	c := Config{
		ListenAddress: "127.0.0.1",
		WebUIPort:     9996,
		MediaPath:     "media/",
		ArchivePath:   "archive.sqlite3",
		ThumbnailPath: "thumb.db",
	}
	c.Logging.FileLogging = false
	c.Logging.LogLevel = log.StringToLogLevel("info").String()
	c.Logging.slogLogLevel = log.StringToLogLevel("info")
	c.Logging.FileLoggingPath = "/logs"
	c.Logging.Profiling = PROFILING_NONE
	c.Debug.DynamicWebReloading = DynamicWebReloading{
		false,
		"",
	}

	return c
}

func Open(path string) (Config, error) {
	var c Config

	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	j := json.NewDecoder(f)
	if err := j.Decode(&c); err != nil {
		return Config{}, err
	}

	c.Logging.slogLogLevel = log.StringToLogLevel(c.Logging.LogLevel)
	return c, nil
}
