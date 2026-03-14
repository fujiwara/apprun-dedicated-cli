package cli

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	"github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runDeploy(ctx context.Context) error {
	if err := c.app.Validate(); err != nil {
		return fmt.Errorf("invalid application definition: %w", err)
	}

	appOp := apprun.NewApplicationOp(c.client)

	// Find or create application
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		slog.Info("application not found, creating", "name", c.app.Name, "cluster", c.app.Cluster)
		clusterID, err := findClusterIDByName(ctx, c.client, c.app.Cluster)
		if err != nil {
			return err
		}
		created, err := appOp.Create(ctx, c.app.Name, clusterID)
		if err != nil {
			return fmt.Errorf("failed to create application: %w", err)
		}
		appDetail, err = appOp.Read(ctx, created.ApplicationID)
		if err != nil {
			return fmt.Errorf("failed to read created application: %w", err)
		}
		slog.Info("created application", "name", c.app.Name, "id", uuid.UUID(appDetail.ApplicationID).String())
	} else {
		slog.Info("deploying application", "name", c.app.Name, "id", uuid.UUID(appDetail.ApplicationID).String())
	}

	// Create new version
	params := definitionToCreateParams(c.app)
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
	slog.Info("created version", "version", ver.GetVersion(), "image", c.app.Image)

	// Activate the new version
	newVer := int32(ver.GetVersion())
	if err := appOp.Update(ctx, appDetail.ApplicationID, &newVer); err != nil {
		return fmt.Errorf("failed to activate version %d: %w", newVer, err)
	}
	slog.Info("activated version", "version", newVer)

	if !c.Deploy.Wait {
		return nil
	}

	// Wait for deployment to complete
	slog.Info("waiting for deployment to complete", "timeout", c.Deploy.WaitTimeout)
	return waitForDeployment(ctx, appOp, appDetail.ApplicationID, newVer, c.Deploy.WaitTimeout)
}

func waitForDeployment(ctx context.Context, appOp *application.ApplicationOp, appID v1.ApplicationID, version int32, timeoutDuration time.Duration) error {
	ticker := time.NewTicker(waitInterval)
	defer ticker.Stop()
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for deployment to complete")
		case <-ticker.C:
			placements, err := appOp.Containers(ctx, appID)
			if err != nil {
				slog.Warn("failed to get containers", "err", err)
				continue
			}

			var running, total int
			for _, p := range placements {
				for _, c := range p.ContainersStats.Containers {
					if c.ApplicationVersion == int64(version) {
						total++
						if c.State == "running" {
							running++
						}
					}
				}
			}

			if total > 0 && running == total {
				slog.Info("deployment complete", "version", version, "running", running)
				return nil
			}
			slog.Info("waiting for containers", "version", version, "running", running, "total", total)
		}
	}
}
