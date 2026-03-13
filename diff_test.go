package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func init() {
	// Disable color output in tests for predictable string matching
	color.NoColor = true
}

func TestPrintColoredDiff_NoDiff(t *testing.T) {
	a := `{"name": "test"}` + "\n"
	var buf bytes.Buffer
	found := printColoredDiff(&buf, a, a, "remote", "local")
	if found {
		t.Errorf("expected no diff, got:\n%s", buf.String())
	}
}

func TestPrintColoredDiff_WithChanges(t *testing.T) {
	a := `{
  "cluster": "default",
  "name": "app",
  "cpu": 1,
  "memory": 1,
  "image": "ghcr.io/example/app:v1.0.0"
}
`
	b := `{
  "cluster": "default",
  "name": "app",
  "cpu": 2,
  "memory": 1,
  "image": "ghcr.io/example/app:v2.0.0"
}
`
	var buf bytes.Buffer
	found := printColoredDiff(&buf, a, b, "remote", "local")
	if !found {
		t.Fatal("expected diff to be found")
	}
	result := buf.String()
	if !strings.Contains(result, "--- remote") {
		t.Error("expected --- remote header")
	}
	if !strings.Contains(result, "+++ local") {
		t.Error("expected +++ local header")
	}
	if !strings.Contains(result, `-  "cpu": 1,`) {
		t.Error("expected deleted cpu line")
	}
	if !strings.Contains(result, `+  "cpu": 2,`) {
		t.Error("expected added cpu line")
	}
	if !strings.Contains(result, `-  "image": "ghcr.io/example/app:v1.0.0"`) {
		t.Error("expected deleted image line")
	}
	if !strings.Contains(result, `+  "image": "ghcr.io/example/app:v2.0.0"`) {
		t.Error("expected added image line")
	}
}

func TestMarshalForDiff(t *testing.T) {
	def := &ApplicationDefinition{
		Cluster: "test",
		Name:    "app",
		CPU:     1,
		Memory:  1,
		Image:   "ghcr.io/example/app:v1.0.0",
	}
	result, err := marshalForDiff(def)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"cluster": "test"`) {
		t.Error("expected cluster field in output")
	}
	if !strings.HasSuffix(result, "\n") {
		t.Error("expected trailing newline")
	}
}
