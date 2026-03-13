package cli

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func init() {
	color.NoColor = true
}

func TestColoredDiff_NoDiff(t *testing.T) {
	result := coloredDiff("")
	if strings.TrimSpace(result) != "" {
		t.Errorf("expected empty diff, got:\n%s", result)
	}
}

func TestColoredDiff_WithChanges(t *testing.T) {
	input := `--- remote
+++ local
-  "cpu": 1,
+  "cpu": 2,
   "memory": 1,`

	result := coloredDiff(input)
	if !strings.Contains(result, `"cpu": 1`) {
		t.Error("expected deleted cpu line")
	}
	if !strings.Contains(result, `"cpu": 2`) {
		t.Error("expected added cpu line")
	}
}

func TestToMap(t *testing.T) {
	def := &ApplicationDefinition{
		Cluster: "test",
		Name:    "app",
		CPU:     1,
		Memory:  1,
		Image:   "ghcr.io/example/app:v1.0.0",
	}
	m := toMap(def)
	if m["cluster"] != "test" {
		t.Errorf("expected cluster=test, got %v", m["cluster"])
	}
	if m["name"] != "app" {
		t.Errorf("expected name=app, got %v", m["name"])
	}
}
