package config

import (
	"encoding/json"
	"os"
)

const (
	PROFILING_CPU = "cpu"
	PROFILING_MEM = "memory"
)

type Config struct {
	Logging struct {
		LogLevel        string
		Profiling       string
		FileLoggingPath string
		FileLogging     bool
	}
	MediaPath   string
	ArchivePath string
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
