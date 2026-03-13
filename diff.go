package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aereal/jsondiff"
	"github.com/fatih/color"
	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runDiff(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	// Get remote (deployed) definition from active version
	var remote *ApplicationDefinition
	if appDetail.ActiveVersion != nil {
		verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
		verDetail, err := verOp.Read(ctx, v1.ApplicationVersionNumber(*appDetail.ActiveVersion))
		if err != nil {
			return fmt.Errorf("failed to read active version: %w", err)
		}
		remote = versionDetailToDefinition(verDetail)
		remote.Cluster = c.app.Cluster
		remote.Name = c.app.Name
	} else {
		remote = &ApplicationDefinition{
			Cluster: c.app.Cluster,
			Name:    c.app.Name,
		}
	}

	remoteID := uuid.UUID(appDetail.ApplicationID).String()
	slog.Info("comparing", "local", c.App, "remote", remoteID)

	diff, err := jsondiff.Diff(
		&jsondiff.Input{Name: remoteID, X: toMap(remote)},
		&jsondiff.Input{Name: c.App, X: toMap(c.app)},
	)
	if err != nil {
		return fmt.Errorf("failed to diff: %w", err)
	}
	if diff == "" {
		fmt.Fprintln(os.Stderr, "No differences found.")
		return nil
	}
	fmt.Print(coloredDiff(diff))
	return nil
}

func coloredDiff(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, "-") {
			b.WriteString(color.RedString(line) + "\n")
		} else if strings.HasPrefix(line, "+") {
			b.WriteString(color.GreenString(line) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	return b.String()
}

func toMap(v any) map[string]any {
	m := make(map[string]any)
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	return m
}
