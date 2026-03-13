package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	armed "github.com/fujiwara/jsonnet-armed"
)

func TestLoadConfig(t *testing.T) {
	ctx := context.Background()
	c := &CLI{
		Config: "testdata/config.jsonnet",
		loader: &armed.CLI{},
	}
	if err := c.setupVM(ctx); err != nil {
		t.Fatal(err)
	}
	if err := c.loadConfig(ctx); err != nil {
		t.Fatal(err)
	}
	cfg := c.config
	if cfg.Cluster != "default" {
		t.Errorf("cluster: got %q, want %q", cfg.Cluster, "default")
	}
	if cfg.Application != "printenv" {
		t.Errorf("application: got %q, want %q", cfg.Application, "printenv")
	}
	if cfg.ApplicationDefinition != "testdata/application.jsonnet" {
		t.Errorf("application_definition: got %q, want %q", cfg.ApplicationDefinition, "testdata/application.jsonnet")
	}
}

func TestLoadConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name:    "missing cluster",
			config:  `{"application": "test", "application_definition": "app.jsonnet"}`,
			wantErr: "config: cluster is required",
		},
		{
			name:    "missing application",
			config:  `{"cluster": "test", "application_definition": "app.jsonnet"}`,
			wantErr: "config: application is required",
		},
		{
			name:    "missing application_definition",
			config:  `{"cluster": "test", "application": "test"}`,
			wantErr: "config: application_definition is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			if err := json.Unmarshal([]byte(tt.config), cfg); err != nil {
				t.Fatal(err)
			}
			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error: got %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoadApplicationDefinition(t *testing.T) {
	ctx := context.Background()
	c := &CLI{
		Config: "testdata/config.jsonnet",
		loader: &armed.CLI{},
		config: &Config{
			ApplicationDefinition: "testdata/application.jsonnet",
		},
	}
	if err := c.setupVM(ctx); err != nil {
		t.Fatal(err)
	}
	def, err := c.loadApplicationDefinition(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if def.CPU != 1 {
		t.Errorf("cpu: got %d, want 1", def.CPU)
	}
	if def.Memory != 1 {
		t.Errorf("memory: got %d, want 1", def.Memory)
	}
	if def.Image.Path != "ghcr.io/fujiwara/printenv" {
		t.Errorf("image.path: got %q, want %q", def.Image.Path, "ghcr.io/fujiwara/printenv")
	}
	if def.Image.Tag != "v0.2.5" {
		t.Errorf("image.tag: got %q, want %q", def.Image.Tag, "v0.2.5")
	}
	if len(def.ExposedPorts) != 1 {
		t.Fatalf("exposed_ports: got %d, want 1", len(def.ExposedPorts))
	}
	if def.ExposedPorts[0].TargetPort != 8080 {
		t.Errorf("exposed_ports[0].target_port: got %d, want 8080", def.ExposedPorts[0].TargetPort)
	}
}

func TestApplicationDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		def     ApplicationDefinition
		wantErr bool
	}{
		{
			name: "valid: lb port with host",
			def: ApplicationDefinition{
				ExposedPorts: []ExposedPort{
					{TargetPort: 8080, LoadBalancerPort: 80, Host: []string{"app.example.com"}},
				},
			},
		},
		{
			name: "valid: no lb port",
			def: ApplicationDefinition{
				ExposedPorts: []ExposedPort{
					{TargetPort: 8080},
				},
			},
		},
		{
			name: "invalid: lb port without host",
			def: ApplicationDefinition{
				ExposedPorts: []ExposedPort{
					{TargetPort: 8080, LoadBalancerPort: 80},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.def.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJsonnetEnvFunction(t *testing.T) {
	ctx := context.Background()
	loader := &armed.CLI{}
	c := &CLI{loader: loader}
	if err := c.setupVM(ctx); err != nil {
		t.Fatal(err)
	}

	t.Setenv("TEST_IMAGE_TAG", "v1.0.0")

	var buf bytes.Buffer
	loader.SetWriter(&buf)
	loader.Filename = "testdata/application.jsonnet"
	if err := loader.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}
