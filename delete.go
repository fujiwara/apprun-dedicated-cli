package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
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

	slog.Info("deleting application", "name", c.app.Name, "id", appID)
	if err := appOp.Delete(ctx, appDetail.ApplicationID); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}
	slog.Info("deleted application", "name", c.app.Name)
	return nil
}
