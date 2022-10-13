package aws

import (
	"context"

	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
)

type AWSCloud struct {
}

func New(ak, sk, regionId string) (*AWSCloud, error) {

	return &AWSCloud{}, nil
}

// ProviderType
func (*AWSCloud) ProviderType() string {
	return cloud.AWSCloud
}

// QueryAccountBill
func (p *AWSCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {

	return types.DataInQueryAccountBill{}, nil
}
