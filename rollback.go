package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runRollback(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	var activeVer *int32
	if appDetail.ActiveVersion != nil {
		v := int32(*appDetail.ActiveVersion)
		activeVer = &v
	}

	var targetVer int32
	if c.Rollback.Target != nil {
		targetVer = *c.Rollback.Target
	} else if activeVer != nil {
		// Find the previous existing version before the active one
		prev, err := findPreviousVersion(ctx, c.client, appDetail.ApplicationID, *activeVer)
		if err != nil {
			return err
		}
		targetVer = prev
	} else {
		// Deactivated: find the latest existing version
		latest, err := findLatestVersion(ctx, c.client, appDetail.ApplicationID)
		if err != nil {
			return err
		}
		targetVer = latest
	}

	if activeVer != nil && targetVer == *activeVer {
		return fmt.Errorf("version %d is already active", targetVer)
	}

	if !c.Rollback.Force {
		var msg string
		if activeVer != nil {
			msg = fmt.Sprintf("Rollback application %q from version %d to %d?", c.app.Name, *activeVer, targetVer)
		} else {
			msg = fmt.Sprintf("Activate version %d for application %q (currently deactivated)?", targetVer, c.app.Name)
		}
		if !confirm(msg) {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	if activeVer != nil {
		slog.Info("rolling back", "from", *activeVer, "to", targetVer)
	} else {
		slog.Info("activating version", "version", targetVer)
	}
	if err := appOp.Update(ctx, appDetail.ApplicationID, &targetVer); err != nil {
		return fmt.Errorf("failed to activate version %d: %w", targetVer, err)
	}
	slog.Info("activated version", "version", targetVer)

	if !c.Rollback.Wait {
		return nil
	}

	slog.Info("waiting for deployment to complete")
	return waitForDeployment(ctx, appOp, appDetail.ApplicationID, targetVer)
}

// findPreviousVersion finds the latest existing version before the given active version.
func findPreviousVersion(ctx context.Context, client *v1.Client, appID v1.ApplicationID, activeVer int32) (int32, error) {
	verOp := apprun.NewVersionOp(client, appID)
	var best int32
	var found bool
	var cursor *v1.ApplicationVersionNumber
	for {
		versions, next, err := verOp.List(ctx, 10, cursor)
		if err != nil {
			return 0, fmt.Errorf("failed to list versions: %w", err)
		}
		for _, v := range versions {
			ver := int32(v.Version)
			if ver < activeVer && (!found || ver > best) {
				best = ver
				found = true
			}
		}
		if next == nil {
			break
		}
		cursor = next
	}
	if !found {
		return 0, fmt.Errorf("no previous version found before version %d", activeVer)
	}
	return best, nil
}

// findLatestVersion finds the latest existing version of the application.
func findLatestVersion(ctx context.Context, client *v1.Client, appID v1.ApplicationID) (int32, error) {
	verOp := apprun.NewVersionOp(client, appID)
	var best int32
	var found bool
	var cursor *v1.ApplicationVersionNumber
	for {
		versions, next, err := verOp.List(ctx, 10, cursor)
		if err != nil {
			return 0, fmt.Errorf("failed to list versions: %w", err)
		}
		for _, v := range versions {
			ver := int32(v.Version)
			if !found || ver > best {
				best = ver
				found = true
			}
		}
		if next == nil {
			break
		}
		cursor = next
	}
	if !found {
		return 0, fmt.Errorf("no versions found")
	}
	return best, nil
}
