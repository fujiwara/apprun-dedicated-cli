package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runAutoScalingGroup(ctx context.Context) error {
	clusterID, err := findClusterIDByName(ctx, c.client, c.app.Cluster)
	if err != nil {
		return err
	}

	op := apprun.NewAutoScalingGroupOp(c.client, clusterID)
	var allASGs []v1.ReadAutoScalingGroupDetail
	var cursor *v1.AutoScalingGroupID
	for {
		asgs, next, err := op.List(ctx, 10, cursor)
		if err != nil {
			return fmt.Errorf("failed to list auto scaling groups: %w", err)
		}
		allASGs = append(allASGs, asgs...)
		if next == nil {
			break
		}
		cursor = next
	}

	if allASGs == nil {
		allASGs = []v1.ReadAutoScalingGroupDetail{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(allASGs)
}
