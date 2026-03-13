package cli

import (
	apprun "github.com/sacloud/apprun-dedicated-api-go"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
)

func newClient(sc *saclient.Client) (*v1.Client, error) {
	return apprun.NewClient(sc)
}
