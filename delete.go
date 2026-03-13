package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	"github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runDelete(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	appID := uuid.UUID(appDetail.ApplicationID).String()
	if !c.Delete.Force {
		fmt.Fprintf(os.Stderr,
			"Are you sure you want to delete application %q (id: %s) in cluster %q? [y/N]: ",
			c.app.Name, appID, c.app.Cluster,
		)
		var answer string
		fmt.Fscanln(os.Stdin, &answer)
		if answer != "y" && answer != "Y" {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	// Deactivate the application before deleting (API requires no active version and no running containers)
	if appDetail.ActiveVersion != nil {
		slog.Info("deactivating application", "name", c.app.Name, "active_version", *appDetail.ActiveVersion)
		if err := appOp.Update(ctx, appDetail.ApplicationID, nil); err != nil {
			return fmt.Errorf("failed to deactivate application: %w", err)
		}
	}

	// Wait for the application to stop running
	slog.Info("waiting for application to stop")
	if err := waitForStopped(ctx, appOp, appDetail.ApplicationID); err != nil {
		return err
	}

	slog.Info("deleting application", "name", c.app.Name, "id", appID)
	if err := appOp.Delete(ctx, appDetail.ApplicationID); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	slog.Info("deleted application", "name", c.app.Name)
	return nil
}

func waitForStopped(ctx context.Context, appOp *application.ApplicationOp, appID v1.ApplicationID) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	timeout := time.After(3 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for application to stop")
		case <-ticker.C:
			detail, err := appOp.Read(ctx, appID)
			if err != nil {
				return fmt.Errorf("failed to read application: %w", err)
			}
			if detail.DesiredCount == nil || *detail.DesiredCount == 0 {
				slog.Info("application stopped")
				return nil
			}
			slog.Info("still running", "desired_count", *detail.DesiredCount)
		}
	}
}
