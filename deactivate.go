package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
)

func (c *CLI) runDeactivate(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	if appDetail.ActiveVersion == nil {
		fmt.Fprintln(os.Stderr, "Application is already deactivated.")
		return nil
	}

	appID := uuid.UUID(appDetail.ApplicationID).String()
	if !c.Deactivate.Force {
		msg := fmt.Sprintf("Are you sure you want to deactivate application %q (id: %s) in cluster %q?",
			c.app.Name, appID, c.app.Cluster)
		if !confirm(msg) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	slog.Info("deactivating application", "name", c.app.Name, "active_version", *appDetail.ActiveVersion)
	if err := appOp.Update(ctx, appDetail.ApplicationID, nil); err != nil {
		return fmt.Errorf("failed to deactivate application: %w", err)
	}

	if !c.Deactivate.Wait {
		slog.Info("deactivated application", "name", c.app.Name)
		return nil
	}

	slog.Info("waiting for application to stop", "timeout", c.Deactivate.WaitTimeout)
	if err := waitForStopped(ctx, appOp, appDetail.ApplicationID, c.Deactivate.WaitTimeout); err != nil {
		return err
	}

	slog.Info("deactivated application", "name", c.app.Name)
	return nil
}
