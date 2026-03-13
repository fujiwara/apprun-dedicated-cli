package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runDeploy(ctx context.Context) error {
	// Load and validate application definition
	def, err := c.loadApplicationDefinition(ctx)
	if err != nil {
		return err
	}
	if err := def.Validate(); err != nil {
		return fmt.Errorf("invalid application definition: %w", err)
	}

	appOp := apprun.NewApplicationOp(c.client)

	// Find or create application
	appDetail, err := findApplicationByName(ctx, appOp, c.config.Cluster, c.config.Application)
	if err != nil {
		slog.Info("application not found, creating", "name", c.config.Application, "cluster", c.config.Cluster)
		clusterID, err := findClusterIDByName(ctx, c.client, c.config.Cluster)
		if err != nil {
			return err
		}
		created, err := appOp.Create(ctx, c.config.Application, clusterID)
		if err != nil {
			return fmt.Errorf("failed to create application: %w", err)
		}
		appDetail, err = appOp.Read(ctx, created.ApplicationID)
		if err != nil {
			return fmt.Errorf("failed to read created application: %w", err)
		}
		slog.Info("created application", "name", c.config.Application, "id", uuid.UUID(appDetail.ApplicationID).String())
	} else {
		slog.Info("deploying application", "name", c.config.Application, "id", uuid.UUID(appDetail.ApplicationID).String())
	}

	// Create new version
	params := definitionToCreateParams(def)
	// For existing apps with versions, keep registry password; for first version, remove
	if params.RegistryPasswordAction == "" {
		if appDetail.ActiveVersion != nil {
			params.RegistryPasswordAction = v1.RegistryPasswordActionKeep
		} else {
			params.RegistryPasswordAction = v1.RegistryPasswordActionRemove
		}
	}
	verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
	ver, err := verOp.Create(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}
	slog.Info("created version", "version", ver.GetVersion(), "image", def.Image.Path+":"+def.Image.Tag)

	// Activate the new version
	newVer := int32(ver.GetVersion())
	if err := appOp.Update(ctx, appDetail.ApplicationID, &newVer); err != nil {
		return fmt.Errorf("failed to activate version %d: %w", newVer, err)
	}
	slog.Info("activated version", "version", newVer)

	return nil
}
