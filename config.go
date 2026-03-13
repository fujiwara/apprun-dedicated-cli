package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
)

// Config represents the CLI configuration file
type Config struct {
	Cluster               string `json:"cluster"`
	Application           string `json:"application"`
	ApplicationDefinition string `json:"application_definition"`
}

func (c *CLI) loadConfig(ctx context.Context) error {
	slog.Info("loading config", "file", c.Config)

	var buf bytes.Buffer
	c.loader.SetWriter(&buf)
	c.loader.Filename = c.Config
	if err := c.loader.Run(ctx); err != nil {
		return fmt.Errorf("failed to evaluate config file %s: %w", c.Config, err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(buf.Bytes(), cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	// Resolve application_definition path relative to config file directory
	if cfg.ApplicationDefinition != "" && !filepath.IsAbs(cfg.ApplicationDefinition) {
		configDir := filepath.Dir(c.Config)
		cfg.ApplicationDefinition = filepath.Join(configDir, cfg.ApplicationDefinition)
	}

	c.config = cfg
	slog.Debug("config loaded", "cluster", cfg.Cluster, "application", cfg.Application)
	return nil
}

func (cfg *Config) Validate() error {
	if cfg.Cluster == "" {
		return fmt.Errorf("config: cluster is required")
	}
	if cfg.Application == "" {
		return fmt.Errorf("config: application is required")
	}
	if cfg.ApplicationDefinition == "" {
		return fmt.Errorf("config: application_definition is required")
	}
	return nil
}
