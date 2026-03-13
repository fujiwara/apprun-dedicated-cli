package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runVersions(ctx context.Context) error {
	appOp := apprun.NewApplicationOp(c.client)
	appDetail, err := findApplicationByName(ctx, appOp, c.app.Cluster, c.app.Name)
	if err != nil {
		return err
	}

	verOp := apprun.NewVersionOp(c.client, appDetail.ApplicationID)
	var allVersions []v1.ApplicationVersionDeploymentStatus
	var cursor *v1.ApplicationVersionNumber
	for {
		versions, next, err := verOp.List(ctx, 10, cursor)
		if err != nil {
			return fmt.Errorf("failed to list versions: %w", err)
		}
		allVersions = append(allVersions, versions...)
		if next == nil {
			break
		}
		cursor = next
	}

	type versionEntry struct {
		Version         int32  `json:"version"`
		Image           string `json:"image"`
		ActiveNodeCount int64  `json:"activeNodeCount"`
		Active          bool   `json:"active"`
		Created         string `json:"created"`
	}

	entries := make([]versionEntry, 0, len(allVersions))
	for _, v := range allVersions {
		active := appDetail.ActiveVersion != nil && int32(*appDetail.ActiveVersion) == int32(v.Version)
		entries = append(entries, versionEntry{
			Version:         int32(v.Version),
			Image:           v.Image,
			ActiveNodeCount: v.ActiveNodeCount,
			Active:          active,
			Created:         time.Unix(int64(v.Created), 0).Format(time.RFC3339),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}
