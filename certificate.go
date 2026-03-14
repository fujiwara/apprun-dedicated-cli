package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

func (c *CLI) runCertificate(ctx context.Context) error {
	clusterID, err := findClusterIDByName(ctx, c.client, c.app.Cluster)
	if err != nil {
		return err
	}

	op := apprun.NewCertificateOp(c.client, clusterID)
	var allCerts []v1.ReadCertificate
	var cursor *v1.CertificateID
	for {
		certs, next, err := op.List(ctx, 10, cursor)
		if err != nil {
			return fmt.Errorf("failed to list certificates: %w", err)
		}
		allCerts = append(allCerts, certs...)
		if next == nil {
			break
		}
		cursor = next
	}

	if allCerts == nil {
		allCerts = []v1.ReadCertificate{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(allCerts)
}
