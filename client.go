package cli

import (
	apprun "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

func newClient(sc *saclient.Client) (*v1.Client, error) {
	return apprun.NewClient(sc)
}
