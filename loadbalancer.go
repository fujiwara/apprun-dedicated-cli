package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
	"github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

type loadBalancerInfo struct {
	LoadBalancerDetail loadbalancer.LoadBalancerDetail       `json:"loadBalancer"`
	Nodes              []loadbalancer.LoadBalancerNodeDetail `json:"nodes"`
}

func (c *CLI) runLoadBalancer(ctx context.Context) error {
	clusterID, err := findClusterIDByName(ctx, c.client, c.app.Cluster)
	if err != nil {
		return err
	}

	asgOp := apprun.NewAutoScalingGroupOp(c.client, clusterID)
	var allLBs []loadBalancerInfo
	var asgCursor *v1.AutoScalingGroupID
	for {
		asgs, next, err := asgOp.List(ctx, 10, asgCursor)
		if err != nil {
			return fmt.Errorf("failed to list auto scaling groups: %w", err)
		}
		for _, asg := range asgs {
			lbOp := apprun.NewLoadBalancerOp(c.client, clusterID, asg.AutoScalingGroupID)
			var lbCursor *v1.LoadBalancerID
			for {
				lbs, lbNext, err := lbOp.List(ctx, 10, lbCursor)
				if err != nil {
					return fmt.Errorf("failed to list load balancers: %w", err)
				}
				for _, lb := range lbs {
					detail, err := lbOp.Read(ctx, lb.LoadBalancerID)
					if err != nil {
						slog.Warn("failed to read load balancer", "id", lb.LoadBalancerID, "err", err)
						continue
					}
					info := loadBalancerInfo{LoadBalancerDetail: *detail}

					// Get node details (contains VIP addresses)
					nodes, err := lbOp.ListNodes(ctx, lb.LoadBalancerID, 10, nil)
					if err != nil {
						slog.Warn("failed to list load balancer nodes", "id", lb.LoadBalancerID, "err", err)
					} else {
						for _, node := range nodes {
							nodeDetail, err := lbOp.ReadNode(ctx, lb.LoadBalancerID, node.LoadBalancerNodeID)
							if err != nil {
								slog.Warn("failed to read load balancer node", "id", node.LoadBalancerNodeID, "err", err)
								continue
							}
							info.Nodes = append(info.Nodes, *nodeDetail)
						}
					}
					allLBs = append(allLBs, info)
				}
				if lbNext == nil {
					break
				}
				lbCursor = lbNext
			}
		}
		if next == nil {
			break
		}
		asgCursor = next
	}

	if allLBs == nil {
		allLBs = []loadBalancerInfo{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(allLBs)
}
