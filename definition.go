package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// ApplicationDefinition mirrors the CreateApplicationVersion API structure.
// JSON field names match the API (camelCase) for forward compatibility.
// Cluster and Name are used to identify the target, not sent to the API.
type ApplicationDefinition struct {
	Cluster           string        `json:"cluster"`
	Name              string        `json:"name"`
	CPU               int64         `json:"cpu"`
	Memory            int64         `json:"memory"`
	ScalingMode       string        `json:"scalingMode,omitempty"`
	FixedScale        *int32        `json:"fixedScale,omitempty"`
	MinScale          *int32        `json:"minScale,omitempty"`
	MaxScale          *int32        `json:"maxScale,omitempty"`
	ScaleInThreshold  *int32        `json:"scaleInThreshold,omitempty"`
	ScaleOutThreshold *int32        `json:"scaleOutThreshold,omitempty"`
	Image             string        `json:"image"`
	Cmd               []string      `json:"cmd,omitempty"`
	RegistryUsername  *string       `json:"registryUsername,omitempty"`
	RegistryPassword  *string       `json:"registryPassword,omitempty"`
	ExposedPorts      []ExposedPort `json:"exposedPorts,omitempty"`
	Env               []EnvVar      `json:"env,omitempty"`
}

type ExposedPort struct {
	TargetPort       int          `json:"targetPort"`
	LoadBalancerPort *int         `json:"loadBalancerPort,omitempty"`
	UseLetsEncrypt   bool         `json:"useLetsEncrypt,omitempty"`
	Host             []string     `json:"host,omitempty"`
	HealthCheck      *HealthCheck `json:"healthCheck,omitempty"`
}

type HealthCheck struct {
	Path            string `json:"path"`
	IntervalSeconds int32  `json:"intervalSeconds,omitempty"`
	TimeoutSeconds  int32  `json:"timeoutSeconds,omitempty"`
}

type EnvVar struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Secret bool   `json:"secret,omitempty"`
}

func (def *ApplicationDefinition) Validate() error {
	if def.Cluster == "" {
		return fmt.Errorf("cluster is required")
	}
	if def.Name == "" {
		return fmt.Errorf("name is required")
	}
	for i, ep := range def.ExposedPorts {
		if ep.LoadBalancerPort != nil && len(ep.Host) == 0 {
			return fmt.Errorf("exposedPorts[%d]: host is required when loadBalancerPort is set", i)
		}
	}
	return nil
}

func (c *CLI) loadApp(ctx context.Context) error {
	slog.Info("loading application definition", "file", c.App)

	var buf bytes.Buffer
	c.loader.SetWriter(&buf)
	c.loader.Filename = c.App
	if err := c.loader.Run(ctx); err != nil {
		return fmt.Errorf("failed to evaluate %s: %w", c.App, err)
	}
	// First try strict decoding to detect unknown fields.
	// Since the input is jsonnet output, it is always valid JSON,
	// so errors here are only caused by unknown fields.
	def := &ApplicationDefinition{}
	strict := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	strict.DisallowUnknownFields()
	if err := strict.Decode(def); err != nil {
		slog.Warn("unknown fields in application definition", "err", err)
		// Re-unmarshal ignoring unknown fields
		def = &ApplicationDefinition{}
		if err := json.Unmarshal(buf.Bytes(), def); err != nil {
			return fmt.Errorf("failed to unmarshal application definition: %w", err)
		}
	}
	if err := def.Validate(); err != nil {
		return fmt.Errorf("invalid application definition: %w", err)
	}
	c.app = def
	return nil
}
