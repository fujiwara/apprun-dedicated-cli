package cli

import (
	"context"

	armed "github.com/fujiwara/jsonnet-armed"
	"github.com/sacloud/saclient-go"
)

func New(ctx context.Context) (*CLI, error) {
	c := &CLI{
		sc:     &saclient.Client{},
		loader: &armed.CLI{},
	}
	return c, nil
}
