package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
)

func (c *CLI) runContainers(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	placements, err := appOp.Containers(ctx, appDetail.ApplicationID)
	if err != nil {
		return fmt.Errorf("failed to get containers: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(placements)
}
