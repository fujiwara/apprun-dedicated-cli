package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// ApplicationDefinition represents the application definition file (deploy target)
type ApplicationDefinition struct {
	CPU                  int                   `json:"cpu"`
	Memory               int                   `json:"memory"`
	ScalingMode          string                `json:"scaling_mode,omitempty"`
	FixedScale           *int                  `json:"fixed_scale,omitempty"`
	MinScale             *int                  `json:"min_scale,omitempty"`
	MaxScale             *int                  `json:"max_scale,omitempty"`
	ScaleInThreshold     *int                  `json:"scale_in_threshold,omitempty"`
	ScaleOutThreshold    *int                  `json:"scale_out_threshold,omitempty"`
	Image                ImageDefinition       `json:"image"`
	Cmd                  []string              `json:"cmd,omitempty"`
	ExposedPorts         []ExposedPort         `json:"exposed_ports,omitempty"`
	EnvironmentVariables []EnvironmentVariable `json:"environment_variables,omitempty"`
}

type ImageDefinition struct {
	Path             string `json:"path"`
	Tag              string `json:"tag"`
	RegistryUsername string `json:"registry_username,omitempty"`
	RegistryPassword string `json:"registry_password,omitempty"`
}

type ExposedPort struct {
	TargetPort       int          `json:"target_port"`
	LoadBalancerPort int          `json:"load_balancer_port,omitempty"`
	UseLetsEncrypt   bool         `json:"use_lets_encrypt,omitempty"`
	Host             []string     `json:"host,omitempty"`
	HealthCheck      *HealthCheck `json:"health_check,omitempty"`
}

type HealthCheck struct {
	Path            string `json:"path"`
	IntervalSeconds int    `json:"interval_seconds,omitempty"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
}

type EnvironmentVariable struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Secret bool   `json:"secret,omitempty"`
}

func (def *ApplicationDefinition) Validate() error {
	for i, ep := range def.ExposedPorts {
		if ep.LoadBalancerPort != 0 && len(ep.Host) == 0 {
			return fmt.Errorf("exposed_ports[%d]: host is required when load_balancer_port is set", i)
		}
	}
	return nil
}

func (c *CLI) loadApplicationDefinition(ctx context.Context) (*ApplicationDefinition, error) {
	name := c.config.ApplicationDefinition
	if name == "" {
		return nil, fmt.Errorf("application_definition is not set in config")
	}
	slog.Info("loading application definition", "file", name)

	var buf bytes.Buffer
	c.loader.SetWriter(&buf)
	c.loader.Filename = name
	if err := c.loader.Run(ctx); err != nil {
		return nil, fmt.Errorf("failed to evaluate %s: %w", name, err)
	}
	def := &ApplicationDefinition{}
	if err := json.Unmarshal(buf.Bytes(), def); err != nil {
		return nil, fmt.Errorf("failed to unmarshal application definition: %w", err)
	}
	return def, nil
}
