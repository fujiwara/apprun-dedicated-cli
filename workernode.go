package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	apprun "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/workernode"
)

func (c *CLI) runWorkerNode(ctx context.Context) error {
	clusterID, err := findClusterIDByName(ctx, c.client, c.app.Cluster)
	if err != nil {
		return err
	}

	asgOp := apprun.NewAutoScalingGroupOp(c.client, clusterID)
	var allNodes []workernode.WorkerNodeDetail
	var asgCursor *v1.AutoScalingGroupID
	for {
		asgs, next, err := asgOp.List(ctx, 10, asgCursor)
		if err != nil {
			return fmt.Errorf("failed to list auto scaling groups: %w", err)
		}
		for _, asg := range asgs {
			nodeOp := apprun.NewWorkerNodeOp(c.client, clusterID, asg.AutoScalingGroupID)
			var nodeCursor *v1.WorkerNodeID
			for {
				nodes, nodeNext, err := nodeOp.List(ctx, 10, nodeCursor)
				if err != nil {
					return fmt.Errorf("failed to list worker nodes: %w", err)
				}
				allNodes = append(allNodes, nodes...)
				if nodeNext == nil {
					break
				}
				nodeCursor = nodeNext
			}
		}
		if next == nil {
			break
		}
		asgCursor = next
	}

	if allNodes == nil {
		allNodes = []workernode.WorkerNodeDetail{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(allNodes)
}
