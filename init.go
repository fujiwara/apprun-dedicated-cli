package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet/formatter"
	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	"github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

const (
	defaultConfigFilename        = "config.jsonnet"
	defaultAppDefinitionFilename = "application.jsonnet"
)

func (c *CLI) runInit(ctx context.Context) error {
	clusterName := c.Init.Cluster
	appName := c.Init.Application
	outputDir := c.Init.OutputDir

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	configFile := filepath.Join(outputDir, defaultConfigFilename)
	appDefFile := filepath.Join(outputDir, defaultAppDefinitionFilename)

	// Find application by name in cluster
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, clusterName, appName)
	if err != nil {
		return err
	}
	slog.Info("found application", "name", appName, "id", uuid.UUID(appDetail.ApplicationID).String())

	// Get active version details
	appDef := &ApplicationDefinition{}
	if appDetail.ActiveVersion != nil {
		verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
		verDetail, err := verOp.Read(ctx, v1.ApplicationVersionNumber(*appDetail.ActiveVersion))
		if err != nil {
			return fmt.Errorf("failed to read active version: %w", err)
		}
		appDef = versionDetailToDefinition(verDetail)
	} else {
		// No active version, try to get latest version
		verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
		vers, _, err := verOp.List(ctx, 1, nil)
		if err == nil && len(vers) > 0 {
			verDetail, err := verOp.Read(ctx, vers[0].GetVersion())
			if err == nil {
				appDef = versionDetailToDefinition(verDetail)
			}
		}
	}

	// Write application definition
	if err := writeJsonnet(appDefFile, appDef); err != nil {
		return err
	}
	slog.Info("wrote application definition", "file", appDefFile)

	// Write config file
	config := &Config{
		Cluster:               clusterName,
		Application:           appName,
		ApplicationDefinition: defaultAppDefinitionFilename,
	}
	if err := writeJsonnet(configFile, config); err != nil {
		return err
	}
	slog.Info("wrote config", "file", configFile)

	return nil
}

func findClusterIDByName(ctx context.Context, client *v1.Client, name string) (v1.ClusterID, error) {
	op := apprun.NewClusterOp(client)
	var cursor *v1.ClusterID
	for {
		clusters, next, err := op.List(ctx, 10, cursor)
		if err != nil {
			return v1.ClusterID{}, fmt.Errorf("failed to list clusters: %w", err)
		}
		for _, c := range clusters {
			if c.Name == name {
				return c.ClusterID, nil
			}
		}
		if next == nil {
			break
		}
		cursor = next
	}
	return v1.ClusterID{}, fmt.Errorf("cluster %q not found", name)
}

func findApplicationByName(ctx context.Context, op *application.ApplicationOp, clusterName, appName string) (*application.ApplicationDetail, error) {
	var cursor *string
	for {
		apps, next, err := op.List(ctx, 10, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to list applications: %w", err)
		}
		for _, a := range apps {
			if a.GetName() == appName {
				detail, err := op.Read(ctx, a.GetApplicationID())
				if err != nil {
					return nil, fmt.Errorf("failed to read application %q: %w", appName, err)
				}
				if detail.ClusterName == clusterName {
					return detail, nil
				}
			}
		}
		if next == nil {
			break
		}
		cursor = next
	}
	return nil, fmt.Errorf("application %q not found in cluster %q", appName, clusterName)
}

func writeJsonnet(filename string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	formatted, err := formatter.Format(filename, string(b), formatter.DefaultOptions())
	if err != nil {
		return fmt.Errorf("failed to format jsonnet: %w", err)
	}
	return os.WriteFile(filename, []byte(formatted), 0644)
}
