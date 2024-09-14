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
	Logging struct {
		LogLevel        string
		Profiling       string
		FileLoggingPath string
		FileLogging     bool
	}
	MediaPath          string
	ArchivePath        string
	ListenAddress      string
	WebUIPort, APIPort int
}

func DefaultValues() Config {
	c := Config{
		ListenAddress: "127.0.0.1",
		WebUIPort:     9996,
		APIPort:       9995,
		MediaPath:     "/media",
		ArchivePath:   "archive.sqlite3",
	}
	c.Logging.FileLogging = false
	c.Logging.LogLevel = "debug"
	c.Logging.FileLoggingPath = "/logs"
	c.Logging.Profiling = PROFILING_NONE

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
