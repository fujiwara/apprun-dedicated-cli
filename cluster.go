package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
)

func (c *CLI) runCluster(ctx context.Context) error {
	clusterID, err := findClusterIDByName(ctx, c.client, c.app.Cluster)
	if err != nil {
		return err
	}
	op := apprun.NewClusterOp(c.client)
	detail, err := op.Read(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to read cluster: %w", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(detail)
}
