package huawei

import (
	"context"

	"github.com/galayx-future/costpilot/internal/constants/cloud"
	"github.com/galayx-future/costpilot/internal/providers/types"
)

type HuaweiCloud struct {
}

func New(ak, sk, regionId string) (*HuaweiCloud, error) {

	return &HuaweiCloud{}, nil
}

// ProviderType
func (*HuaweiCloud) ProviderType() string {
	return cloud.HuaweiCloud
}

// QueryAccountBill
func (p *HuaweiCloud) QueryAccountBill(ctx context.Context, param types.QueryAccountBillRequest) (types.DataInQueryAccountBill, error) {

	return types.DataInQueryAccountBill{}, nil
}
