package cli

import (
	"bytes"
	"context"
	"testing"

	armed "github.com/fujiwara/jsonnet-armed"
)

func TestLoadApp(t *testing.T) {
	ctx := context.Background()
	c := &CLI{
		App:    "testdata/application.jsonnet",
		loader: &armed.CLI{},
	}
	if err := c.setupVM(ctx); err != nil {
		t.Fatal(err)
	}
	if err := c.loadApp(ctx); err != nil {
		t.Fatal(err)
	}
	def := c.app
	if def.Cluster != "default" {
		t.Errorf("cluster: got %q, want %q", def.Cluster, "default")
	}
	if def.Name != "printenv" {
		t.Errorf("name: got %q, want %q", def.Name, "printenv")
	}
	if def.CPU != 1 {
		t.Errorf("cpu: got %d, want 1", def.CPU)
	}
	if def.Memory != 1 {
		t.Errorf("memory: got %d, want 1", def.Memory)
	}
	if def.Image != "ghcr.io/fujiwara/printenv:v0.2.5" {
		t.Errorf("image: got %q, want %q", def.Image, "ghcr.io/fujiwara/printenv:v0.2.5")
	}
	if len(def.ExposedPorts) != 1 {
		t.Fatalf("exposedPorts: got %d, want 1", len(def.ExposedPorts))
	}
	if def.ExposedPorts[0].TargetPort != 8080 {
		t.Errorf("exposedPorts[0].targetPort: got %d, want 8080", def.ExposedPorts[0].TargetPort)
	}
	if def.ExposedPorts[0].LoadBalancerPort == nil || *def.ExposedPorts[0].LoadBalancerPort != 80 {
		t.Errorf("exposedPorts[0].loadBalancerPort: got %v, want 80", def.ExposedPorts[0].LoadBalancerPort)
	}
}

func TestApplicationDefinition_Validate(t *testing.T) {
	lbPort := 80
	tests := []struct {
		name    string
		def     ApplicationDefinition
		wantErr bool
	}{
		{
			name: "valid",
			def: ApplicationDefinition{
				Cluster: "test",
				Name:    "app",
				ExposedPorts: []ExposedPort{
					{TargetPort: 8080, LoadBalancerPort: &lbPort, Host: []string{"app.example.com"}},
				},
			},
		},
		{
			name: "valid: no lb port",
			def: ApplicationDefinition{
				Cluster: "test",
				Name:    "app",
				ExposedPorts: []ExposedPort{
					{TargetPort: 8080},
				},
			},
		},
		{
			name: "invalid: missing cluster",
			def: ApplicationDefinition{
				Name: "app",
			},
			wantErr: true,
		},
		{
			name: "invalid: missing name",
			def: ApplicationDefinition{
				Cluster: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid: lb port without host",
			def: ApplicationDefinition{
				Cluster: "test",
				Name:    "app",
				ExposedPorts: []ExposedPort{
					{TargetPort: 8080, LoadBalancerPort: &lbPort},
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
