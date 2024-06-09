package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	EnableDebug        bool
	EnableFileLogging  bool
	EnableCPUProfiling bool
	EnableMemProfiling bool
	mediaPath          string
	archivePath        string
	loggingPath        string
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

func (c Config) MediaPath() string {
	return c.mediaPath
}
func (c Config) ArchivePath() string {
	return c.archivePath
}

func (c Config) LoggingPath() string {
	return c.loggingPath
}
