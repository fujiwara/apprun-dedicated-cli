package cli

import (
	"context"
	"fmt"
	"log/slog"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runRollback(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	if appDetail.ActiveVersion == nil {
		return fmt.Errorf("no active version to rollback from")
	}
	activeVer := int32(*appDetail.ActiveVersion)

	var targetVer int32
	if c.Rollback.Target != nil {
		targetVer = *c.Rollback.Target
	} else {
		// Find the previous existing version before the active one
		prev, err := findPreviousVersion(ctx, c.client, appDetail.ApplicationID, activeVer)
		if err != nil {
			return err
		}
		targetVer = prev
	}

	if targetVer == activeVer {
		return fmt.Errorf("version %d is already active", targetVer)
	}

	slog.Info("rolling back", "from", activeVer, "to", targetVer)
	if err := appOp.Update(ctx, appDetail.ApplicationID, &targetVer); err != nil {
		return fmt.Errorf("failed to activate version %d: %w", targetVer, err)
	}
	slog.Info("activated version", "version", targetVer)
	return nil
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
