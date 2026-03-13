package cli

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/google/uuid"
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runStatus(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.config.Cluster, c.config.Application)
	if err != nil {
		return err
	}

	status := map[string]any{
		"cluster": c.config.Cluster,
		"application": map[string]any{
			"id":             uuid.UUID(appDetail.ApplicationID).String(),
			"name":           appDetail.Name,
			"active_version": appDetail.ActiveVersion,
			"desired_count":  appDetail.DesiredCount,
		},
	}

	// Get version details if active
	if appDetail.ActiveVersion != nil {
		verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
		verDetail, err := verOp.Read(ctx, v1.ApplicationVersionNumber(*appDetail.ActiveVersion))
		if err != nil {
			slog.Warn("failed to read active version", "err", err)
		} else {
			status["active_version_detail"] = map[string]any{
				"version":           verDetail.Version,
				"image":             verDetail.Image,
				"cpu":               verDetail.CPU,
				"memory":            verDetail.Memory,
				"scaling_mode":      verDetail.ScalingMode,
				"active_node_count": verDetail.ActiveNodeCount,
			}
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(status)
}

func (c *CLI) runRender(ctx context.Context) error {
	def, err := c.loadApplicationDefinition(ctx)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(def)
}
