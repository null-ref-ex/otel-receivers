package shellreceiver

import (
	"fmt"
	"os"
	"time"
)

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	Interval string `mapstructure:"interval"`
	Path     string `mapstructure:"path"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	interval, _ := time.ParseDuration(cfg.Interval)
	if interval.Minutes() < 1 {
		return fmt.Errorf("when defined, the interval has to be set to at least 1 minute (1m)")
	}

	if _, err := os.Stat(cfg.Path); err != nil {
		return fmt.Errorf("File does not exist\n")
	}
	return nil
}
