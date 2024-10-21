package config

import (
	"encoding/json"
	"os"
)

const (
	PROFILING_CPU  = "cpu"
	PROFILING_MEM  = "memory"
	PROFILING_NONE = "none"
)

type Config struct {
	Debug struct {
		DynamicWebReloading bool
	}
	Logging struct {
		LogLevel        string
		Profiling       string
		FileLoggingPath string
		FileLogging     bool
	}
	MediaPath          string
	ArchivePath        string
	ThumbnailPath      string
	ListenAddress      string
	WebUIPort, APIPort int
}

// DefaultValues returns a config with sane defaults
func DefaultValues() Config {
	c := Config{
		ListenAddress: "127.0.0.1",
		WebUIPort:     9996,
		APIPort:       9995,
		MediaPath:     "/media",
		ArchivePath:   "archive.sqlite3",
		ThumbnailPath: "thumb.db",
	}
	c.Logging.FileLogging = false
	c.Logging.LogLevel = "info"
	c.Logging.FileLoggingPath = "/logs"
	c.Logging.Profiling = PROFILING_NONE
	c.Debug = struct{ DynamicWebReloading bool }{
		DynamicWebReloading: false,
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

	return c, nil
}
